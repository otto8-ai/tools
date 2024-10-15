import path from 'node:path'
import { search } from './search.ts'
import { refineQuery } from './refineQuery.ts'
import { getNewContext } from './context.ts'
import * as gptscript from '@gptscript-ai/gptscript'
import { rmSync } from 'node:fs'

const gptsClient = new gptscript.GPTScript()

const query: string = process.env.QUERY ?? ''
if (query === '') {
  console.log('error: no query provided')
  process.exit(1)
}

if (process.env.GPTSCRIPT_WORKSPACE_ID === undefined || process.env.GPTSCRIPT_WORKSPACE_DIR === undefined) {
  console.log('error: GPTScript workspace ID and directory are not set')
  process.exit(1)
}

// Simultaneously start the browser and generate our search query.
const [refinedQuery, context, noJSContext] = await Promise.all([
  refineQuery(query),
  getNewContext(path.resolve(process.env.GPTSCRIPT_WORKSPACE_DIR), true),
  getNewContext(path.resolve(process.env.GPTSCRIPT_WORKSPACE_DIR), false)
])

// Query Google
const pageContents = await search(
  context.context,
  noJSContext.context,
  refinedQuery,
  process.env.MAXRESULTS !== undefined ? parseInt(process.env.MAXRESULTS) : undefined
)

// Ask gpt-4o-mini to generate an answer
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
  context: [],
  credentials: [],
  description: '',
  export: [],
  exportContext: [],
  globalModelName: '',
  globalTools: [],
  jsonResponse: true,
  maxTokens: 0,
  modelName: 'gpt-4o-mini',
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

// Each session dir is usually at least 20MB, and they are one-time use, so we don't need to keep them around.
await context.context.close()
await noJSContext.context.close()
rmSync(context.sessionDir, { recursive: true, force: true })
rmSync(noJSContext.sessionDir, { recursive: true, force: true })
process.stdout.write(JSON.stringify(focusedSearch))

process.exit(0)

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
