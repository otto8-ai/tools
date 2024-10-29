import { BrowserContext } from '@playwright/test';
import * as cheerio from 'cheerio';
import { getMarkdown } from './getMarkdown.ts';

export type SearchResult = {
  url: string;
  content: string;
}

export async function search(
  context: BrowserContext,
  query: string,
  maxResults: number = 3 // Default to 3 results if not specified
): Promise<SearchResult[]> {
  const foundURLs = new Set<string>()
  const contentsPromises: Array<Promise<SearchResult | null>> = []

  const page = await context.newPage();
  const noJSPages = await Promise.all(
    Array.from({ length: maxResults }, async () => {
      const page = await context.newPage()
      await page.addInitScript(() => {
        // Disable JavaScript for the page
        Object.defineProperty(navigator, 'javaScriptEnabled', { value: false });
        Object.defineProperty(window, 'Function', { value: () => {} });
        Object.defineProperty(window, 'eval', { value: () => {} });
      })

      return page
    })
  )

  try {
    await page.goto(`https://www.google.com/search?q=${query}&udm=14`);
    const content = await page.content();
    const $ = cheerio.load(content);
    const elements = $('#rso a[jsname]');
  
    elements.each((_, element) => {
      if (contentsPromises.length >= maxResults) return false;

      const url = $(element).attr('href') ?? '';
      if (url && !url.includes('youtube.com/watch?v') && !foundURLs.has(url)) {
        foundURLs.add(url);
        contentsPromises.push(getMarkdown(noJSPages[contentsPromises.length], url).then(content => {
          return content ? { url, content } : null;
        }));
      }
    });

    const results = (await Promise.all(contentsPromises)).filter(Boolean) as SearchResult[];
    return results;
  } finally {
    await page.close();
    await Promise.all(noJSPages.map(p => p.close()));
  }
}
