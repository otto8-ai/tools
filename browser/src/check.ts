import { type Page } from 'playwright'
import { inspect } from './browse.ts'
import { delay } from './delay.ts'

// check checks a checkbox or radio button.
export async function check (page: Page, model: string, userInput: string, keywords: string[], matchTextOnly: boolean): Promise<void> {
  const locators = await inspect(page, model, userInput, 'check', matchTextOnly, keywords)
  console.log(locators)
  for (const locator of locators) {
    try {
      await page.check(`${locator}`, { timeout: 5000 })
      await delay(3000)
      break
    } catch (e) {
      try {
        await page.check(`[for=${locator}]`, { timeout: 5000 })
        break
      } catch (e) {
      }
    }
  }
}
