import express, {type Request, type Response} from "express"
import bodyParser from "body-parser"
import {type Page} from "playwright"
import {browse, filterContent} from "./browse.ts"
import {fill} from "./fill.ts"
import {enter} from "./enter.ts"
import {scrollToBottom} from "./scrollToBottom.ts"
import {randomBytes} from "node:crypto"
import {getSessionId, SessionManager} from "./session.ts"
import {screenshot, ScreenshotInfo} from "./screenshot.ts"

async function main (): Promise<void> {
  console.log('Starting browser server')

  const app = express()
  // Get port from the environment variable or use 9888 if it is not defined
  const port = parseInt(process.env.PORT ?? "9888", 10)
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

    const model: string = data.model ?? process.env.OBOT_DEFAULT_LLM_MINI_MODEL ?? 'gpt-4o-mini'
    const website: string = data.website ?? ''
    const userInput: string = data.userInput ?? ''
    const keywords: string[] = (data.keywords ?? '').split(',')
    const filter: string = data.filter ?? ''
    const followMode: boolean = data.followMode === 'false' ? false : Boolean(data.followMode)

    try {
      const sessionID = getSessionId(req.headers)
      await sessionManager.withSession(sessionID, async (browserContext, openPages) => {
        const tabID = data.tabID ?? randomBytes(8).toString('hex')
        const printTabID = data.tabID === undefined
        let takeScreenshot = followMode

        // Get the page for this tab, creating a new one if it doesn't exist or the existing page is closed
        let page: Page = openPages.get(tabID)!
        if (!openPages.has(tabID) || page.isClosed()) {
            page = await browserContext.newPage()
            openPages.set(tabID, page)
        }
        await page.bringToFront()

        let response: { result?: any, screenshotInfo?: ScreenshotInfo } = {}
        switch (req.path) {
          case '/browse':
            // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
            response.result = await browse(page, website, 'browse', tabID, printTabID)
            break

          case '/getFilteredContent':
            response.result = await filterContent(page, tabID, printTabID, filter)
            break

          case '/getPageContents':
            response.result = await browse(page, website, 'getPageContents', tabID, printTabID)
            break

          case '/getPageLinks':
            response.result = await browse(page, website, 'getPageLinks', tabID, printTabID)
            break

          case '/getPageImages':
            response.result = await browse(page, website, 'getPageImages', tabID, printTabID)
            break

          case '/fill':
            // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
            response.result = await fill(page, model, userInput, data.content ?? '', keywords, (data.matchTextOnly as boolean) ?? false)
            break

          case '/enter':
            await enter(page)
            break

          case '/scrollToBottom':
            await scrollToBottom(page)
            break

          case '/screenshot':
            takeScreenshot = true
            break

          case '/back':
            await page.goBack()
            break

          case '/forward':
            await page.goForward()
            break

          default:
            throw new Error(`Unknown tool endpoint: ${req.path}`)
        }

        if (takeScreenshot) {
          const fullPage = data.fullPage === 'false' ? false : Boolean(data.fullPage)
          response.screenshotInfo = await screenshot(page, req.headers, tabID, fullPage)
        }

        res.json(response)
      })
    } catch (e) {
      // Send a 200 status code GPTScript will pass the error to the LLM
      res.status(200).send(`Error: ${e}`)
    } finally {
      res.end()
    }
  })

  // Start the server
  const server = app.listen(port, "127.0.0.1", () => {
    console.log(`Server is listening on port ${port}`)
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

// eslint-disable-next-line @typescript-eslint/no-floating-promises
await main()
