import { GPTScript, type ToolDef } from '@gptscript-ai/gptscript'
import { type SearchResults, type SearchResult } from './search.ts'

const gptscript = new GPTScript()

export async function refine (model: string, unrefined: SearchResults): Promise<SearchResults> {
  const now = new Date().toISOString()
  const refined = await Promise.all(
    unrefined.results.map(async content => await refineResult(model, now, unrefined.query, content))
  )

  return {
    ...unrefined,
    results: refined.filter(result => hasContent(result.content))
  }
}

function hasContent (content?: string | string[]): boolean {
  return !(Array.isArray(content) ? content?.length === 0 : content?.trim() === '')
}

async function refineResult (
  model: string,
  time: string,
  query: string,
  result: SearchResult): Promise<SearchResult> {
  const tool: ToolDef = {
    chat: false,
    jsonResponse: true,
    modelName: model,
    temperature: 0.0,
    arguments: {
      type: 'object',
      properties: {
        time: {
          type: 'string',
          description: 'Current date and time that the search was requested at'
        },
        query: {
          type: 'string',
          description: 'query or subject matter to generate citations for'
        },
        url: {
          type: 'string',
          description: 'URL that the content was sourced from'
        },
        content: {
          type: 'string',
          description: 'Markdown content to cite'
        }
      },
      required: ['query', 'url', 'content']
    },
    instructions: refineInstructions
  }

  const run = await gptscript.evaluate(tool, {
    input: JSON.stringify({
      query,
      ...result,
      time
    })
  })

  return await run.json()
}

// Note: Tools can't introspect their parameters schema, so we provide it in the instructions as well
const refineInstructions = `
Given an object with the following JSON schema:

${minify({
  type: 'object',
  properties: {
    time: {
      type: 'string',
      description: 'Current date and time that the search was requested at'
    },
    query: {
      type: 'string',
      description: 'Query or subject matter to generate citations for'
    },
    url: {
      type: 'string',
      description: 'URL that the content was sourced from'
    },
    content: {
      type: 'string',
      description: 'Markdown content to cite'
    }
  },
  required: ['query', 'url', 'content', 'time']
})}

Select all markdown from \${CONTENT} containing information useful to cite when researching \${QUERY}.
Selected markdown should contain the most useful and relevant information to \${QUERY} available in \${CONTENT}.
Don't select markdown that is not helpful or related to \${QUERY}.
 
Respond with a single object containing all of the selected markdown that adheres to the following JSON schema:

${minify({
  type: 'object',
  properties: {
    url: {
      type: 'string',
      description: 'URL that the content was sourced from'
    },
    title: {
      type: 'string',
      description: 'Main title of the source content'
    },
    content: {
      type: 'array',
      description: 'Cleaned up markdown from the original content that can be cited to research the query',
      items: {
        type: 'string'
      }
    }
  },
  required: ['url', 'title', 'content']
})}

Do not respond with any additional dialog or commentary.
`

function minify (obj: object): string {
  return JSON.stringify(obj).replace(/\n/g, '')
}
