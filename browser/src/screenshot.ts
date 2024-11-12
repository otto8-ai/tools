import { type Page } from 'playwright'
import * as gptscript from '@gptscript-ai/gptscript'
import { getWorkspaceId } from './session.ts'
import { type IncomingHttpHeaders } from 'node:http'

const client = new gptscript.GPTScript();

export async function screenshot (page: Page, headers: IncomingHttpHeaders): Promise<string> {
  const workspaceId = getWorkspaceId(headers['x-gptscript-env'])
  let screenshotName = `screenshot-${Date.now()}_${page.url().replace(/[^a-zA-Z0-9]/g, '_')}.png`
  // Detect if we are running in otto8
  if (process.env.OTTO8_THREAD_ID !== undefined) {
    screenshotName = `files/${screenshotName}`
  }
  const screenshot = await page.screenshot()

  try {
    await client.writeFileInWorkspace(screenshotName, screenshot, workspaceId)
  } catch (err) {
    console.error(err)
  }
  return JSON.stringify({ results: {workspaceLocation: screenshotName }})
}
