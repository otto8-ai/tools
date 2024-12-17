import { type Page } from 'playwright'
import { GPTScript } from '@gptscript-ai/gptscript'
import { getWorkspaceId, getGPTScriptEnv } from './session.ts'
import { type IncomingHttpHeaders } from 'node:http'
import { createHash } from 'node:crypto'

const client = new GPTScript()

export interface ScreenshotInfo {
  tabID: string
  tabPageUrl: string
  takenAt: number
  imageWorkspaceFile: string
  imageDownloadUrl: string | undefined
}

export async function screenshot (
  page: Page,
  headers: IncomingHttpHeaders,
  tabID: string,
  fullPage: boolean = false): Promise<ScreenshotInfo> {
  // Generate a unique workspace file name for the screenshot
  const timestamp = Date.now()
  const pageHash = createHash('sha256').update(page.url()).digest('hex').substring(0, 8)
  const screenshotName = `screenshot-${timestamp}_${pageHash}.png`

  try {
    // Take the screenshot
    const screenshot = await page.screenshot({ fullPage, animations: 'disabled' })

    // If we are running in obot, we need to save the screenshot in the files directory
    const workspaceId = getWorkspaceId(headers)
    const screenshotPath = workspaceId !== undefined ? `files/${screenshotName}` : screenshotName

    // Save the screenshot to the workspace
    await client.writeFileInWorkspace(screenshotPath, screenshot, workspaceId)
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    throw new Error(`Failed to save screenshot to workspace: ${msg}`)
  }

  // Build the download URL used by the UI to display the image
  let downloadUrl: string | undefined
  const obotServerUrl = getGPTScriptEnv(headers, 'OBOT_SERVER_URL')
  const threadId = getGPTScriptEnv(headers, 'OBOT_THREAD_ID')
  if (obotServerUrl !== undefined && threadId !== undefined) {
    downloadUrl = `${obotServerUrl}/api/threads/${threadId}/file/${screenshotName}`
  }

  return {
    tabID,
    tabPageUrl: page.url(),
    takenAt: timestamp,
    imageWorkspaceFile: screenshotName,
    imageDownloadUrl: downloadUrl
  }
}
