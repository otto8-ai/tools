import { type Page } from 'playwright'
import { inspect } from './browse.ts'
import path from 'node:path'

export async function screenshot (page: Page, model: string, userInput: string, keywords: string[], filename: string, matchTextOnly: boolean): Promise<void> {
  const locators = await inspect(page, model, userInput, 'screenshot', matchTextOnly, keywords)
  let done = false
  for (const locator of locators) {
    console.log(locator)
    try {
      const elements = await page.locator('css=' + locator).all()
      if (elements.length === 0) {
        console.log('no elements found for locator ' + locator)
        continue
      }

      await elements[0].screenshot({ path: path.resolve(process.env.GPTSCRIPT_WORKSPACE_DIR + '/' + filename) })
      done = true
      break
    } catch (e) {
      console.error('failed to take screenshot: ', e)
    }
  }

  if (!done) {
    console.error('failed to take screenshot')
  }
}
