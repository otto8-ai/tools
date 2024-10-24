import { type Page } from 'playwright'
import { delay } from './delay.ts'

export async function scrollToBottom (page: Page): Promise<void> {
  await page.keyboard.press('End')
  await delay(2000)
}
