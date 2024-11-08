import { type BrowserContext, type Page } from '@playwright/test'
import * as cheerio from 'cheerio'
import TurndownService from 'turndown'

export interface SearchResult {
  url: string
  title?: string
  content?: string | string[]
}

export interface SearchResults {
  query: string
  results: SearchResult[]
}

export async function search (
  context: BrowserContext,
  query: string,
  maxResults: number
): Promise<SearchResults> {
  if (query === '') {
    throw new Error('No query provided')
  }
  const foundURLs = new Set<string>()
  const results: Array<Promise<SearchResult | null>> = []

  const page = await context.newPage()
  const noJSPages = await Promise.all(
    Array.from({ length: maxResults }, async () => {
      const page = await context.newPage()
      await page.addInitScript(() => {
        // Disable JavaScript for the page
        Object.defineProperty(navigator, 'javaScriptEnabled', { value: false })
        Object.defineProperty(window, 'Function', { value: () => { } })
        Object.defineProperty(window, 'eval', { value: () => { } })
      })

      return page
    })
  )

  try {
    await page.goto(`https://www.google.com/search?q=${query}&udm=14`)
    const content = await page.content()
    const $ = cheerio.load(content)
    const elements = $('#rso a[jsname]')

    elements.each((_, element) => {
      if (results.length >= maxResults) return false

      const url = $(element).attr('href') ?? ''
      if ((url !== '') && !url.includes('youtube.com/watch?v') && !foundURLs.has(url)) {
        foundURLs.add(url)
        results.push(getMarkdown(noJSPages[results.length], url).then(content => {
          return (content !== '') ? { url, content } : null
        }))
      }
    })

    return {
      query,
      results: (await Promise.all(results)).filter(Boolean) as SearchResult[]
    }
  } finally {
    // Fire and forget page close so we can move on
    void page.close()
    void Promise.all(noJSPages.map(async p => { await p.close() }))
  }
}

export async function getMarkdown (page: Page, url: string): Promise<string> {
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
