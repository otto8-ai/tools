import { type BrowserContext, chromium, firefox } from 'playwright'

export async function newBrowserContext (sessionDir: string, javaScriptEnabled: boolean = true): Promise<BrowserContext> {
  let context: BrowserContext
  const browser = await getSystemBrowser()
  switch (browser) {
    case 'chromium':
      context = await chromium.launchPersistentContext(
        sessionDir,
        {
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
          headless: true,
          viewport: null,
          args: ['--start-maximized', '--disable-blink-features=AutomationControlled'],
          ignoreDefaultArgs: ['--enable-automation'],
          javaScriptEnabled
        })
      break
    case 'chrome':
      context = await chromium.launchPersistentContext(
        sessionDir,
        {
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
          headless: true,
          viewport: null,
          channel: 'chrome',
          args: ['--start-maximized', '--disable-blink-features=AutomationControlled'],
          ignoreDefaultArgs: ['--enable-automation'],
          javaScriptEnabled
        })
      break
    case 'firefox':
      context = await firefox.launchPersistentContext(
        sessionDir,
        {
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Firefox/89.0 Safari/537.36',
          headless: true,
          viewport: null,
          javaScriptEnabled
        })
      break
    case 'edge':
      context = await chromium.launchPersistentContext(
        sessionDir,
        {
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.64',
          headless: true,
          viewport: null,
          channel: 'msedge',
          args: ['--start-maximized', '--disable-blink-features=AutomationControlled'],
          ignoreDefaultArgs: ['--enable-automation'],
          javaScriptEnabled
        })
      break
    default:
      throw new Error(`Unknown browser: ${browser}`)
  }

  return context
}

let systemBrowser: string | undefined

async function getSystemBrowser (): Promise<string> {
  if (systemBrowser != null) {
    return systemBrowser
  }

  const browsers = [
    { name: 'Chrome', launchFunction: async () => await chromium.launch({ channel: 'chrome', headless: true }) },
    { name: 'Edge', launchFunction: async () => await chromium.launch({ channel: 'msedge', headless: true }) },
    { name: 'Firefox', launchFunction: async () => await firefox.launch({ headless: true }) },
    { name: 'Chromium', launchFunction: async () => await chromium.launch({ headless: true }) }
  ]

  const errors = []
  for (const browser of browsers) {
    try {
      const browserInstance = await browser.launchFunction()
      void browserInstance.close()

      systemBrowser = browser.name.toLowerCase()

      return browser.name.toLowerCase()
    } catch (error) {
      errors.push(error)
    }
  }

  throw new Error(`No supported browsers (Chrome, Edge, Firefox) are installed: ${JSON.stringify(errors)}`)
}
