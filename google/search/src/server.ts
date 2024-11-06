import path from 'node:path'
import { search } from './search.ts'
import { refineQuery } from './refineQuery.ts'
import bodyParser from 'body-parser'
import { getSessionId, SessionManager } from './session.ts'
import express, { type Request, type Response } from 'express'
import * as gptscript from '@gptscript-ai/gptscript'

const gptsClient = new gptscript.GPTScript()

function dedent(str: string): string {
  const lines = str.split('\n');
  const indentLength = Math.min(
    ...lines.filter(line => line.trim()).map(line => line.match(/^\s*/)![0].length)
  );
  return lines.map(line => line.slice(indentLength)).join('\n').trim();
}

function minify(obj: object): string {
  return JSON.stringify(obj).replace(/\n/g, '')
}

async function main (): Promise<void> {
  const app = express()

  // Get port from the environment variable or use 9888 if it is not defined
  const port = process.env.PORT ?? 9888
  delete (process.env.GPTSCRIPT_INPUT)
  app.use(bodyParser.json())

  // Start the server
  const server = app.listen(port, () => {
    console.log(`Server is listening on port ${port}`)
  })

  const sessionManager = await SessionManager.create()

  // gptscript requires "GET /" to return 200 status code
  app.get('/', (_req: Request, res: Response) => {
    res.send('OK')
  })

  app.post('/*', async (req: Request, res: Response) => {

    try {
      const data = req.body
      const maxResults = Number.isInteger(parseInt(data.maxResults)) ? parseInt(data.maxResults) : undefined
      const model = data.model || 'gpt-4o-mini'
      const query = data.query || ''
      if (query === '') {
        throw new Error('No query provided')
      }

      const sessionID = getSessionId(req.headers)

      await sessionManager.withSession(sessionID, async (browserContext) => {
        // Query Google
        const refinedQuery = await refineQuery(query)
        const pageContents = await search(
          browserContext,
          refinedQuery,
          maxResults,
        )

        // Ask the LLM to narrow down the search pages to the most relevant content
        const tool: gptscript.ToolDef = {
          agents: [],
          arguments: { 
            type: 'object', 
            properties: {
              search: {
                type: 'string',
                description: 'JSON string containing unfocussed search query and results'
              }
            },
            required: ['search']
          },
          chat: false,
          context: ['github.com/otto8-ai/tools/time'],
          credentials: [],
          description: '',
          export: [],
          exportContext: [],
          globalModelName: '',
          globalTools: [],
          jsonResponse: true,
          maxTokens: 0,
          modelName: model,
          modelProvider: false,
          name: '',
          tools: [],
          temperature: 0.2,
          instructions: dedent(`
            Given a search object with the following JSON schema: 

            ${minify({
              "title": "Search",
              "type": "object",
              "properties": {
                "query": {
                  "type": "string",
                  "description": "The search query used"
                },
                "results": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "properties": {
                      "url": {
                        "type": "string",
                        "format": "uri",
                        "description": "URL of the web page"
                      },
                      "content": {
                        "type": "string",
                        "description": "Content of the web page at URL in markdown"
                      }
                    },
                    "required": ["url", "content"],
                    "additionalProperties": false
                  },
                  "description": "An array of search results"
                }
              },
              "required": ["query", "results"],
              "additionalProperties": false
            })}

            For all search results, silently select all the chunks of text from each result's content that: 
            - answer the search query directly or provide additional information to support an answer to the search query
            - contain enough context from the source text to complete the thought expressed by them in the original source text

            Then generate a single FocusedSearch object containing all of the selected chunks using the following JSON schema:

            ${minify({
              "title": "FocusedSearch",
              "type": "object",
              "properties": {
                "query": {
                  "type": "string",
                  "description": "The search query used"
                },
                "results": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "properties": {
                      "url": {
                        "type": "string",
                        "format": "uri",
                        "description": "URL of the search result's web page"
                      },
                      "title": {
                        "type": "string",
                        "description": "Title of the search result's web page"
                      },
                      "relevantContent": {
                        "type": "array",
                        "items": {
                          "type": "string"
                        },
                        "description": "Chunks of text sourced from the web page at URL that answer the search query"
                      }
                    },
                    "required": ["url", "title", "relevantContent"],
                  },
                  "description": "An array of focused search results"
                }
              },
              "required": ["query", "results"],
            })}

            Remove all results with no relevant content and deduplicate semantically equivalent relevant content before
            responding with the minified JSON of the FocusedSearch object.

            Do not include any additional preamble or commentary.
          `)
        }

        const run = await gptsClient.evaluate(tool, { 
          input: JSON.stringify({ refinedQuery, results: pageContents }),
          disableCache: true 
        })
        const focusedSearch = await run.json()

        res.status(200).send(focusedSearch)
      })
    } catch (e) {
      // Send a 200 status code GPTScript will pass the error to the LLM
      res.status(200).send(`Error: ${e}`)
    } finally {
      res.end()
    }

  })


  // stdin is used as a keep-alive mechanism. When the parent process dies the stdin will be closed and this process
  // will exit.
  process.stdin.resume()
  process.stdin.on('close', () => {
    console.log('Closing the server')
    server.close()
    process.exit(0)
  })

  process.on('SIGINT', () => {
    console.log('Closing the server')
    server.close()
    process.exit(0)
  })

  process.on('SIGTERM', () => {
    console.log('Closing the server')
    server.close()
    process.exit(0)
  })

  process.on('SIGHUP', () => {
    console.log('Closing the server')
    server.close()
    process.exit(0)
  })
}

await main()