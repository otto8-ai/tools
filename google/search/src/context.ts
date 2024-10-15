import { type BrowserContext, chromium } from '@playwright/test'
import { randomInt } from 'node:crypto'

export interface ContextAndSessionDir {
  context: BrowserContext
  sessionDir: string
}

export async function getNewContext (workspaceDir: string, javaScriptEnabled: boolean): Promise<ContextAndSessionDir> {
  const sessionDir = workspaceDir + '/afti_browser_session_' + randomInt(1, 1000000).toString()
  let context: BrowserContext

  context = await chromium.launchPersistentContext(
    sessionDir,
    {
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
      headless: true,
      viewport: null,
      args: ['--start-maximized', '--disable-blink-features=AutomationControlled'],
      ignoreDefaultArgs: ['--enable-automation', '--use-mock-keychain'],
      javaScriptEnabled
    })

  return { context, sessionDir }
}
