import { delay } from './delay.ts'
import { type BrowserContext } from 'playwright'

// login waits for the user to close the browser.
export async function login (context: BrowserContext, website: string): Promise<void> {
  const page = await context.newPage()

  // Listen for the 'close' event on the browser
  page.on('close', () => {
    context.close().then().catch((err) => { console.log(err) })
  })
  await page.goto(website)
  await delay(50000)
}
