import express, { type Request, type Response } from 'express'
import bodyParser from 'body-parser'
import { type Page } from 'playwright'
import { browse, close, filterContent } from './browse.ts'
import { click } from './click.ts'
import { fill } from './fill.ts'
import { enter } from './enter.ts'
import { check } from './check.ts'
import { select } from './select.ts'
import { login } from './login.ts'
import { scrollToBottom } from './scrollToBottom.ts'
import { randomBytes } from 'node:crypto'
import { screenshot } from './screenshot.ts'
import { getSessionId, SessionManager } from './session.ts'

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

    const model: string = data.model ?? 'gpt-4o-mini'
    const website: string = data.website ?? ''
    const userInput: string = data.userInput ?? ''
    const keywords: string[] = (data.keywords ?? '').split(',')
    const filter: string = data.filter ?? ''

    if (process.env.GPTSCRIPT_WORKSPACE_ID === undefined || process.env.GPTSCRIPT_WORKSPACE_DIR === undefined) {
      res.status(400).send('GPTScript workspace ID and directory are not set')
      return
    }

    const sessionID = getSessionId(req.headers)
    await sessionManager.withSession(sessionID, async (browserContext, openPages) => {
      let tabID = randomBytes(8).toString('hex')
      let printTabID = true
      if (data.tabID !== undefined) {
        tabID = data.tabID
        printTabID = false
      }

      browserContext.on('close', () => {
        console.log('Closing the context')
        setTimeout(() => {
          process.exit(0)
        }, 3000)
      })

      let page: Page
      if (openPages[tabID] !== undefined) {
        page = openPages[tabID]
        if (page.isClosed()) {
          page = await browserContext.newPage()
          if (openPages === undefined) {
            openPages = {}
          }
          openPages[tabID] = page
        }
      } else {
        page = await browserContext.newPage()
        if (openPages === undefined) {
          openPages = {}
        }
        openPages[tabID] = page
      }
      await page.bringToFront()

      let allElements = false
      if (data.allElements === 'true' || data.allElements === true) {
        allElements = true
      }

      if (req.path === '/browse') {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
        res.send(await browse(page, website, 'browse', tabID, printTabID))
      } else if (req.path === '/getFilteredContent') {
        res.send(await filterContent(page, tabID, printTabID, filter))
      } else if (req.path === '/getPageContents') {
        res.send(await browse(page, website, 'getPageContents', tabID, printTabID))
      } else if (req.path === '/getPageLinks') {
        res.send(await browse(page, website, 'getPageLinks', tabID, printTabID))
      } else if (req.path === '/getPageImages') {
        res.send(await browse(page, website, 'getPageImages', tabID, printTabID))
      } else if (req.path === '/click') {
        await click(page, model, userInput, keywords.map((keyword) => keyword.trim()), allElements, (data.matchTextOnly as boolean) ?? false)
      } else if (req.path === '/fill') {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
        await fill(page, model, userInput, data.content ?? '', keywords, (data.matchTextOnly as boolean) ?? false)
      } else if (req.path === '/enter') {
        await enter(page)
      } else if (req.path === '/check') {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
        await check(page, model, userInput, keywords, (data.matchTextOnly as boolean) ?? false)
      } else if (req.path === '/select') {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
        await select(page, model, userInput, data.option ?? '')
      } else if (req.path === '/login') {
        await login(browserContext, website)
      } else if (req.path === '/scrollToBottom') {
        await scrollToBottom(page)
      } else if (req.path === '/close') {
        await close(page)
        // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
        delete openPages[tabID]
      } else if (req.path === '/back') {
        await page.goBack()
      } else if (req.path === '/forward') {
        await page.goForward()
      } else if (req.path === '/screenshot') {
        await screenshot(page, model, userInput, keywords, (data.filename as string) ?? 'screenshot.png', (data.matchTextOnly as boolean) ?? false)
      }
    })

    res.end()
  })

  // Start the server
  const server = app.listen(port, () => {
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
