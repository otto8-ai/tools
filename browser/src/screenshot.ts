import { type Page } from 'playwright'
import * as gptscript from '@gptscript-ai/gptscript'
import { getWorkspaceId, getGPTScriptEnv } from './session.ts'
import { type IncomingHttpHeaders } from 'node:http'

const client = new gptscript.GPTScript();
const ottoServerUrl = process.env.OTTO_SERVER_URL

export async function screenshot (page: Page, headers: IncomingHttpHeaders, tabID: string, fullPage: boolean = false): Promise<string> {
  // Detect if we are running in otto8
  const screenshot = await page.screenshot({fullPage, animations: 'disabled'})

  const timestamp = Date.now()
  let screenshotName = `screenshot-${timestamp}_${page.url().replace(/[^a-zA-Z0-9]/g, '_')}.png`
  try {
    // If we are running in otto8, we need to save the screenshot in the files directory
    const workspaceId = getWorkspaceId(headers)
    const screenshotPath = workspaceId !== undefined ? `files/${screenshotName}` : screenshotName
    await client.writeFileInWorkspace(screenshotPath, screenshot, workspaceId)
  } catch (err) {
    console.error(err)
    throw new Error(`Failed to save screenshot to workspace`)
  }

  let downloadUrl: string | undefined
  if (ottoServerUrl !== undefined) {
    const threadId = getGPTScriptEnv(headers, 'OTTO_THREAD_ID')
    downloadUrl = `${ottoServerUrl}/api/threads/${threadId}/file/${screenshotName}`
  }

  return JSON.stringify({screenshotInfo: {
    tabID: tabID,
    tabPageUrl: page.url(),
    takenAt: timestamp,
    imageWorkspaceFile: screenshotName,
    imageDownloadUrl: downloadUrl,
  }})
}
