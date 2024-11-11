import bodyParser from 'body-parser'
import { getSessionId, SessionManager } from './session.ts'
import express, { type Request, type Response, type RequestHandler } from 'express'
import { search } from './search.ts'
import { refine } from './refine.ts'

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

  app.post('/*', (async (req: Request, res: Response): Promise<void> => {
    try {
      const responseStart = performance.now()

      const data = req.body
      const maxResults = Number.isInteger(Number(data.maxResults)) ? parseInt(data.maxResults as string, 10) : 3
      const query: string = data.query ?? ''
      const sessionID = getSessionId(req.headers)

      await sessionManager.withSession(sessionID, async (browserContext) => {
        // Query Google and get the result pages as markdown
        const searchResults = await search(
          browserContext,
          query,
          maxResults
        )
        const searchEnd = performance.now()

        // Extract the relevant citations from the content of each page
        const refinedResults = await refine(searchResults)
        const refineEnd = performance.now()

        res.status(200).send(JSON.stringify({
          duration: {
            search: (searchEnd - responseStart) / 1000,
            refine: (refineEnd - searchEnd) / 1000,
            response: (refineEnd - responseStart) / 1000
          },
          ...refinedResults
        }))
      })
    } catch (error: unknown) {
      const msg = error instanceof Error ? error.message : String(error)
      // Send a 200 status code GPTScript will pass the error to the LLM
      res.status(200).send(`Error: ${msg}`)
    } finally {
      res.end()
    }
  }) as RequestHandler)

  let stopped = false
  const stop = (): void => {
    if (stopped) return
    stopped = true
    console.error('Daemon shutting down...')
    server.close(() => process.exit(0))
  }

  // stdin is used as a keep-alive mechanism
  // When the parent process dies the stdin will be closed and this process
  process.stdin.resume()
  process.stdin.on('close', stop)
  const signals = ['SIGINT', 'SIGTERM', 'SIGHUP']
  signals.forEach(signal => process.on(signal, stop))
}

await main()
