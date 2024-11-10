import express, { type Request, type Response } from 'express'
import bodyParser from 'body-parser'
import { type Page } from 'playwright'
import { browse } from './browse.ts'
// import { browse, filterContent } from './browse.ts'
// import { fill } from './fill.ts'
// import { enter } from './enter.ts'
// import { scrollToBottom } from './scrollToBottom.ts'
import { randomBytes } from 'node:crypto'
import { getSessionId, SessionManager } from './session.ts'
// import { screenshot } from './screenshot.ts'

async function main (): Promise<void> {
  console.log('Starting browser server')

  const app = express()
  // Get port from the environment variable or use 9888 if it is not defined
  const port = process.env.PORT ?? 9888
  delete (process.env.GPTSCRIPT_INPUT)
  app.use(bodyParser.json())

  const sessionManager = await SessionManager.create()

  // gptscript requires "GET /" to return 200 status code
  app.get('/', (_req: Request, res: Response) => {
    res.send('OK')
  })

  // eslint-disable-next-line @typescript-eslint/no-misused-promises
  app.post('/*', async (req: Request, res: Response) => {
    const data = req.body

    // const model: string = data.model ?? 'gpt-4o-mini'
    const website: string = data.website ?? ''
    // const userInput: string = data.userInput ?? ''
    // const keywords: string[] = (data.keywords ?? '').split(',')
    // const filter: string = data.filter ?? ''

    try {
      if (process.env.GPTSCRIPT_WORKSPACE_ID === undefined) {
        throw new Error('GPTScript workspace ID is not set')
      }

      const sessionID = getSessionId(req.headers)
      await sessionManager.withSession(sessionID, async (browserContext, openPages) => {
        let tabID = randomBytes(8).toString('hex')
        let printTabID = true
        if (data.tabID !== undefined) {
          tabID = data.tabID
          printTabID = false
        }

        let page: Page
        if (openPages.has(tabID)) {
          page = openPages.get(tabID)!
          if (page.isClosed()) {
            page = await browserContext.newPage()
            openPages.set(tabID, page)
          }
        } else {
          page = await browserContext.newPage()
          openPages.set(tabID, page)
        }
        await page.bringToFront()

        try {
        switch (req.path) {
          // case '/browse':
          //   // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
          //   res.send(await browse(page, website, 'browse', tabID, printTabID))
          //   break

          // case '/getFilteredContent':
          //   res.send(await filterContent(page, tabID, printTabID, filter))
          //   break

          case '/getPageContents':
            res.send(await browse(page, website, 'getPageContents', tabID, printTabID))
            break

          // case '/getPageLinks':
          //   res.send(await browse(page, website, 'getPageLinks', tabID, printTabID))
          //   break

          // case '/getPageImages':
          //   res.send(await browse(page, website, 'getPageImages', tabID, printTabID))
          //   break

          // case '/fill':
          //   // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
          //   await fill(page, model, userInput, data.content ?? '', keywords, (data.matchTextOnly as boolean) ?? false)
          //   break

          // case '/enter':
          //   await enter(page)
          //   break

          // case '/scrollToBottom':
          //   await scrollToBottom(page)
          //   break

          // case '/screenshot':
          //   res.send(await screenshot(page, req.headers))
          //   break

          // case '/back':
          //   await page.goBack()
          //   break

          // case '/forward':
          //   await page.goForward()
          //   break

          default:
            throw new Error(`Unknown tool endpoint: ${req.path}`)
        }
      } finally {
        // TODO: This is a hack to disable persistent tabs while `Get Page Contents` is the only
        // tool exposed by the Browser bundle.
        // Remove this block when we reintroduce the other browser tools.
        openPages.delete(tabID)
        await page.close()
      }

      })
    } catch (e) {
      // Send a 200 status code GPTScript will pass the error to the LLM
      res.status(200).send(`Error: ${e}`)
    } finally {
      res.end()
    }
  })

  // Start the server
  const server = app.listen(port, () => {
    console.log(`Server is listening on port ${port}`)
  })

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

void await main()
