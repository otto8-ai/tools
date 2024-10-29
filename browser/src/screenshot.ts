import { type Page } from 'playwright'
import * as gptscript from '@gptscript-ai/gptscript'

const client = new gptscript.GPTScript();

export async function screenshot (page: Page): Promise<string> {
  let screenshotName = `screenshot-${Date.now()}_${page.url().replace(/[^a-zA-Z0-9]/g, '_')}.png`
  // Detect if we are running in otto8
  if (process.env.OTTO_THREAD_ID !== undefined) {
    screenshotName = `files/${screenshotName}`
  }
  const screenshot = await page.screenshot()

  try {
    await client.writeFileInWorkspace(screenshotName, screenshot)
  } catch (err) {
    console.error(err)
  }
  return JSON.stringify({ results: {workspaceLocation: screenshotName }})
}
