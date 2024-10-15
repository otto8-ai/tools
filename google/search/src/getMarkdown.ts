import { type Page } from '@playwright/test'
import TurndownService from 'turndown'
import * as cheerio from 'cheerio'

/**
 * Get the markdown content of a page.
 * @param page - The page to get the content from.
 * @param url - The URL of the page.
 * @returns The markdown content of the page.
 */
export async function getMarkdown(page: Page, url: string): Promise<string> {
  try {
    await page.goto(url, { timeout: 1000 })
  } catch (e) {
    console.warn('slow page:', url)
  }

  let content = ''
  while (content === '') {
    let fails = 0
    try {
      content = await page.content()
    } catch (e) {
      fails++
      if (fails > 2) {
        void page.close()
        console.warn('rip:', url)
        return '' // Page didn't load; just ignore.
      }
      await new Promise(resolve => setTimeout(resolve, 100)) // sleep 100ms
    }
  }
  void page.close()

  const $ = cheerio.load(content)

  $('noscript').remove()
  $('script').remove()
  $('style').remove()
  $('img').remove()
  $('g').remove()
  $('svg').remove()
  $('iframe').remove()

  let resp = ''
  const turndownService = new TurndownService()
  $('body').each(function () {
    resp += turndownService.turndown($.html(this))
  })

  return trunc(resp, 80000)
}

function trunc (text: string, max: number): string {
  return text.length > max ? text.slice(0, max) + '...' : text
}
