import { delay } from './delay.ts'
import { type Page } from 'playwright'
import { inspect } from './browse.ts'

export async function fill (page: Page, model: string, userInput: string, content: string, keywords: string[], matchTextOnly: boolean): Promise<void> {
  const locators = await inspect(page, model, userInput, 'fill', matchTextOnly, keywords)
  for (const locator of locators) {
    console.log(locator)
    try {
      const elements = await page.locator('css=' + locator).all()
      if (elements.length === 0) {
        continue
      }

      await elements[0].fill(content)
      await delay(5000)
      break
    } catch (e) {
      console.error(e)
    }
  }
}
