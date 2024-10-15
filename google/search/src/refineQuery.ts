import * as gptscript from '@gptscript-ai/gptscript'

export async function refineQuery(query: string): Promise<string> {
  const tool: gptscript.ToolDef = {
    agents: [],
    arguments: { type: 'object' },
    chat: false,
    context: [],
    credentials: [],
    description: '',
    export: [],
    exportContext: [],
    globalModelName: '',
    globalTools: [],
    jsonResponse: false,
    maxTokens: 0,
    modelProvider: false,
    name: '',
    tools: [],
    modelName: 'gpt-4o-mini',
    instructions: `
    Refine the query below to improve its Google Search results.
    The refined query should preserve the inferred intent of the original query.
    Do not quote the output.

    Query: ${query}
    ` 
  }

  const client = new gptscript.GPTScript()
  const run = await client.evaluate(tool)
  return await run.text()
}
