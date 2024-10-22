import { type Page } from 'playwright'
import * as cheerio from 'cheerio'
import { GPTScript } from '@gptscript-ai/gptscript'
import { delay } from './delay.ts'
import { URL } from 'url'
import TurndownService from 'turndown'

export async function close (page: Page): Promise<void> {
  await page.close()
}

// browse navigates to the website and returns the text content of the page (if print is true)
export async function browse (page: Page, website: string, mode: string, tabID: string, printTabID: boolean): Promise<string> {
  if (website !== '' && page.url() !== website) {
    try {
      await page.goto(website)
    } catch (e) {} // choke the exception because we don't care
    await delay(5000)
  } else {
    await delay(5000)
  }

  const iframes = page.locator('body').locator('iframe')
  const count = await iframes.count()
  for (let i = 0; i < count; i++) {
    const frame = iframes.nth(i)
    if (await frame.isVisible()) {
      try {
        await frame.scrollIntoViewIfNeeded({ timeout: 3000 })
      } catch (e) {} // ignore timeouts
    }
  }

  let resp: string = ''
  if (mode === 'getPageContents') {
    const html = await getPageHTML(page)
    const $ = cheerio.load(html)

    $('script').each(function () {
      const elem = $(this)
      elem.contents().filter(function () {
        return this.type === 'text'
      }).remove()
      const children = elem.contents()
      elem.before(children)
      elem.remove()
    })
    $('noscript').remove()
    $('style').remove()
    $('img').remove()
    $('[style]').removeAttr('style')
    $('[onclick]').removeAttr('onclick')
    $('[onload]').removeAttr('onload')
    $('[onerror]').removeAttr('onerror')

    // Remove empty divs and spans
    $('div').each(function () {
      if ($(this).text() === '' && $(this).children().length === 0) {
        $(this).remove()
      }
    })
    $('span').each(function () {
      if ($(this).text() === '' && $(this).children().length === 0) {
        $(this).remove()
      }
    })

    const turndownService = new TurndownService()
    $('body').each(function () {
      resp += turndownService.turndown($.html(this))
    })
  } else if (mode === 'getPageLinks') {
    const html = await getPageHTML(page)
    const $ = cheerio.load(html)
    $('a').each(function () {
      const link = new URL($(this).attr('href') ?? '', page.url()).toString()
      resp += `[${$(this).text().trim()}](${link.trim()})\n`
    })
  } else if (mode === 'getPageImages') {
    const html = await getPageHTML(page)
    const $ = cheerio.load(html)
    $('img').each(function () {
      resp += `${Object.entries($(this).attr() ?? '').toString()}\n`
    })
  }

  if (printTabID) {
    resp = `Tab ID: ${tabID}\n` + resp
  }
  return resp.split('\n').filter(line => line.trim() !== '').join('\n')
}

export async function filterContent (page: Page, tabID: string, printTabID: boolean, filter: string): Promise<string> {
  // Navigate and get the page contents
  const html = await getPageHTML(page)

  const $ = cheerio.load(html)
  let filteredContent = ''

  // Check if the filter is an ID, class, or tag selector
  if (filter.startsWith('#') || filter.startsWith('.') || /^[a-zA-Z]/.test(filter)) {
    // Use the filter to select elements matching the CSS selector
    $(filter).each((i, elem) => {
      filteredContent += $.html(elem)
    })
  } else {
    throw new Error(`Invalid filter format: ${filter}. Use a CSS ID, class, or tag selector.`)
  }

  // Clean up the filtered content
  const clean$ = cheerio.load(filteredContent)
  clean$('script').each(function () {
    const elem = $(this)
    // Remove text nodes inside the script tag
    elem.contents().filter(function () {
      return this.type === 'text'
    }).remove()
    // Extract the remaining content from the script tag
    const children = elem.contents()
    elem.before(children)
    elem.remove()
  })
  clean$('noscript, style, img').remove()
  clean$('[style]').removeAttr('style')
  clean$('[onclick], [onload], [onerror]').removeAttr('onclick onload onerror')

  // Remove empty divs and spans
  $('div').each(function () {
    if ($(this).text() === '' && $(this).children().length === 0) {
      $(this).remove()
    }
  })
  $('span').each(function () {
    if ($(this).text() === '' && $(this).children().length === 0) {
      $(this).remove()
    }
  })

  // Remove HTML comment nodes
  clean$('*').contents().filter(function () {
    return this.type === 'comment'
  }).remove()

  let resp = clean$.html().trim()
  if (printTabID) {
    resp = `Tab ID: ${tabID}\n` + clean$.html().trim()
  }

  return resp.split('\n').filter(line => line.trim() !== '').join('\n')
}

// inspect inspects a webpage and returns a locator for a specific element based on the user's input and action.
export async function inspect (page: Page, model: string, userInput: string, action: string, matchTextOnly: boolean, keywords?: string[]): Promise<string[]> {
  if (userInput === '') {
    // This shouldn't happen, since the LLM is told that it must fill in the user input,
    // but in case it doesn't, just use the keywords.
    userInput = keywords?.join(' ') ?? ''
  }

  let elementData = ''
  const modes = ['matchAll', 'oneSibling', 'twoSiblings', 'parent', 'grandparent', 'matchAny']
  // TODO - improve this so that we stop faster
  if (matchTextOnly) {
    const elementDataPromises = modes.map(async mode => await summarize(page, keywords ?? [], action, mode, true))
    const results = await Promise.all(elementDataPromises)
    for (const result of results) {
      if (result !== '') {
        elementData = result
        break
      }
    }
  }

  if (elementData === '') {
    // Do it again, but don't match text only
    const elementDataPromises = modes.map(async mode => await summarize(page, keywords ?? [], action, mode, false))
    const results = await Promise.all(elementDataPromises)
    for (const result of results) {
      if (result !== '') {
        elementData = result
        break
      }
    }
  }

  if (elementData === '') {
    // Do it again, but split the keywords by space
    const elementDataPromises = modes.map(async mode => await summarize(page, keywords?.join(' ')?.split(' ') ?? [], action, mode, true))
    const results = await Promise.all(elementDataPromises)
    for (const result of results) {
      if (result !== '') {
        elementData = result
        break
      }
    }
  }

  // Scroll the page to get more data and try to find the element
  // Retry 10 times also to give user time to sign in if the landing page requires sign in
  let retry = 0
  while (elementData === '' && retry < 10) {
    await page.evaluate(() => {
      const allElements = Array.from(document.querySelectorAll('div'))
      const scrollables = allElements.filter(e => ((e.getAttribute('style')?.includes('overflow: auto') ?? false) || (e.getAttribute('style')?.includes('overflow: scroll') ?? false)))
      if (scrollables.length > 0) {
        scrollables[0].scrollBy(0, window.innerHeight)
      }
    })
    await delay(2000)
    elementData = await summarize(page, keywords ?? [], action, 'matchAny', false)
    retry++
  }

  // Safeguard to try to avoid breaking the context window
  if (elementData.length > 200000) {
    elementData = elementData.substring(0, 200000)
  }

  console.log('elementData:', elementData)

  const instructions = getActionInstructions(action, { userInput, elementData })
  const client = new GPTScript()
  const run = await client.evaluate({
    modelName: model,
    instructions
  })

  const output = (await run.text()).replace('\n', '').trim()
  return [output]
}

export async function inspectForSelect (page: Page, model: string, userInput: string, userSelection: string, keywords?: string[]): Promise<{ selector: string, option: string }> {
  let elementData = ''
  const modes = ['matchAll', 'oneSibling', 'twoSiblings', 'parent', 'grandparent']
  for (const mode of modes) {
    elementData = await summarize(page, keywords ?? [], 'select', mode, false)
    if (elementData !== '') {
      break
    }
  }
  if (elementData === '') {
    // Do it again, but split the keywords by space
    for (const mode of modes) {
      elementData = await summarize(page, keywords?.join(' ')?.split(' ') ?? [], 'select', mode, false)
      if (elementData !== '') {
        break
      }
    }
  }

  let retry = 0
  while (elementData === '' && retry < 10) {
    await page.evaluate(() => {
      const allElements = Array.from(document.querySelectorAll('div'))
      const scrollables = allElements.filter(e => ((e.getAttribute('style')?.includes('overflow: auto') ?? false) || (e.getAttribute('style')?.includes('overflow: scroll') ?? false)))
      if (scrollables.length > 0) {
        scrollables[0].scrollBy(0, window.innerHeight)
      }
    })
    await delay(2000)
    elementData = await summarize(page, keywords ?? [], 'select', 'matchAny', false)
    retry++
  }

  const instructions = getActionInstructions('select', { userInput, userSelection, elementData })
  const client = new GPTScript()
  const run = await client.evaluate({
    modelName: model,
    instructions
  })

  const output = (await run.text()).replace('\n', '').trim()
  console.log(output)

  return JSON.parse(output)
}

function matchKeywords ($: cheerio.CheerioAPI, e: cheerio.Element, keywords: string[], mode: string, textOnly: boolean): string | null {
  let html = ''
  let text = ''
  switch (mode) {
    case 'matchAll':
      if (textOnly && keywords.every(keyword => $(e).text().toLowerCase().includes(keyword.toLowerCase()) || $(e).attr('value')?.toLowerCase().includes(keyword.toLowerCase()))) {
        return $.html(e)
      } else if (!textOnly && keywords.every(keyword => $.html(e).toLowerCase().includes(keyword.toLowerCase()))) {
        return $.html(e)
      }
      break
    case 'oneSibling':
      if (textOnly) {
        if (e.prev !== null) {
          text += $(e.prev).text()
          html += $.html(e.prev)
        }
        text += $(e).text()
        html += $.html(e)
        if (e.next !== null) {
          text += $(e.next).text()
          html += $.html(e.next)
        }
        if (keywords.every(keyword => text.toLowerCase().includes(keyword.toLowerCase()) || $(e).attr('value')?.toLowerCase().includes(keyword.toLowerCase()))) {
          return html
        }
      } else {
        if (e.prev !== null) {
          html += $.html(e.prev)
        }
        html += $.html(e)
        if (e.next !== null) {
          html += $.html(e.next)
        }
        if (keywords.every(keyword => html.toLowerCase().includes(keyword.toLowerCase()))) {
          return html
        }
      }
      break
    case 'twoSiblings':
      if (textOnly) {
        if (e.prev?.prev !== null) {
          text += $(e.prev?.prev).text()
          html += $.html(e.prev?.prev)
        }
        if (e.prev !== null) {
          text += $(e.prev).text()
          html += $.html(e.prev)
        }
        text += $(e).text()
        html += $.html(e)
        if (e.next !== null) {
          text += $(e.next).text()
          html += $.html(e.next)
        }
        if (e.next?.next !== null) {
          text += $(e.next?.next).text()
          html += $.html(e.next?.next)
        }
        if (keywords.every(keyword => text.toLowerCase().includes(keyword.toLowerCase()) || $(e).attr('value')?.toLowerCase().includes(keyword.toLowerCase()))) {
          return html
        }
      } else {
        if (e.prev?.prev !== null) {
          html += $.html(e.prev?.prev)
        }
        if (e.prev !== null) {
          html += $.html(e.prev)
        }
        html += $.html(e)
        if (e.next !== null) {
          html += $.html(e.next)
        }
        if (e.next?.next !== null) {
          html += $.html(e.next?.next)
        }
        if (keywords.every(keyword => html.toLowerCase().includes(keyword.toLowerCase()))) {
          return html
        }
      }
      break
    case 'parent':
      if (e.parent !== null) {
        text = $(e.parent).text()
        html = $.html(e.parent)
        if (textOnly && keywords.every(keyword => text.toLowerCase().includes(keyword.toLowerCase()) || $(e).attr('value')?.toLowerCase().includes(keyword.toLowerCase()))) {
          return html
        } else if (!textOnly && keywords.every(keyword => html.toLowerCase().includes(keyword.toLowerCase()))) {
          return html
        }
      }
      break
    case 'grandparent':
      if (e.parent?.parent !== null) {
        text = $(e.parent?.parent).text()
        html = $.html(e.parent?.parent)
        if (textOnly && keywords.every(keyword => text.toLowerCase().includes(keyword.toLowerCase()) || $(e).attr('value')?.toLowerCase().includes(keyword.toLowerCase()))) {
          return html
        } else if (!textOnly && keywords.every(keyword => html.toLowerCase().includes(keyword.toLowerCase()))) {
          return html
        }
      }
      break
    case 'matchAny':
      if (textOnly && keywords.some(keyword => $(e).text().toLowerCase().includes(keyword.toLowerCase()) || $(e).attr('value')?.toLowerCase().includes(keyword.toLowerCase()))) {
        return $.html(e)
      } else if (keywords.some(keyword => $.html(e).toLowerCase().includes(keyword.toLowerCase()))) {
        return $.html(e)
      }
      break
  }

  return null
}

// summarize returns relevant HTML elements for the given keywords and action
export async function summarize (page: Page, keywords: string[], action: string, mode: string, matchTextOnly: boolean): Promise<string> {
  const htmlContent = await getPageHTML(page)
  const $ = cheerio.load(htmlContent)

  $('noscript').remove()
  $('style').remove()
  $('[style]').removeAttr('style')

  let resp = ''
  // For search, we need to find all the input and textarea elements and figure that out
  if (action === 'fill') {
    $('textarea').each(function () {
      if (keywords.length !== 0) {
        const res = matchKeywords($, this, keywords, mode, matchTextOnly)
        if (res !== null && res !== '') {
          resp += res + '\n\n'
        }
      } else {
        resp += $.html(this) + '\n\n'
      }
    })
    $('input[id]').each(function () {
      if (keywords.length !== 0) {
        const res = matchKeywords($, this, keywords, mode, matchTextOnly)
        if (res !== null && res !== '') {
          resp += res + '\n\n'
        }
      } else {
        resp += $.html(this) + '\n\n'
      }
    })
    $('input').each(function () {
      if (keywords.length !== 0) {
        const res = matchKeywords($, this, keywords, mode, matchTextOnly)
        if (res !== null && res !== '') {
          resp += res + '\n\n'
        }
      } else {
        resp += $.html(this) + '\n\n'
      }
    })
    $('form').each(function () {
      if (keywords.length !== 0) {
        const res = matchKeywords($, this, keywords, mode, matchTextOnly)
        if (res !== null && res !== '') {
          resp += res + '\n\n'
        }
      } else {
        resp += $.html(this) + '\n\n'
      }
    })
    $('div').each(function () {
      if (keywords.length !== 0) {
        if ($(this).attr('contenteditable') === 'true') {
          const res = matchKeywords($, this, keywords, mode, matchTextOnly)
          if (res !== null && res !== '') {
            resp += res + '\n\n'
          }
        }
      }
    })
  }

  if (action === 'check') {
    $('input[type="checkbox"]').parents().each(function () {
      const res = matchKeywords($, this, keywords, mode, matchTextOnly)
      if (res !== null && res !== '') {
        resp += res + '\n\n'
      }
    })
  }

  if (action === 'select') {
    $('select').each(function () {
      resp += $.html(this) + '\n\n'
    })
  }

  if (action === 'screenshot') {
    // For screenshot, we want to remove scripts, text-only elements, styles, and other junk
    // so that we can hand the more structural elements to the LLM for it to determine the best locator.
    $('script').remove()
    $('p').remove()
    $('h1').remove()
    $('h2').remove()
    $('h3').remove()
    $('h4').remove()
    $('h5').remove()
    $('li').remove()
    $('a').remove()
    $('[onclick]').removeAttr('onclick')
    $('[onload]').removeAttr('onload')
    $('[onerror]').removeAttr('onerror')
    $('body').each(function () {
      resp += findKeywordsInElement($, this, keywords, mode === 'matchAll') + '\n\n'
    })
  }

  if (action === 'click') {
    $('a').each(function () {
      if (keywords.length !== 0) {
        const res = matchKeywords($, this, keywords, mode, matchTextOnly)
        if (res !== null && res !== '') {
          resp += res + '\n\n'
        }
      } else {
        resp += $.html(this) + '\n\n'
      }
    })
    $('button').each(function () {
      if (keywords.length !== 0) {
        const res = matchKeywords($, this, keywords, mode, matchTextOnly)
        if (res !== null && res !== '') {
          resp += res + '\n\n'
        }
      } else {
        resp += $.html(this) + '\n\n'
      }
    })

    $('div').each(function () {
      if (keywords.length !== 0) {
        if (hasNoNonTextChildren(this)) {
          const res = matchKeywords($, this, keywords, mode, matchTextOnly)
          if (res !== null && res !== '') {
            resp += res + '\n\n'
          }
        }
      } else {
        resp += '<div '
        for (const attr of this.attributes) {
          resp += ` ${attr.name}="${attr.value}"`
        }
        resp += '>' + '\n'
        for (const c of this.children) {
          resp += $.html(c) + '\n'
        }
        resp += '</div>' + '\n\n'
      }
    })

    $('span').each(function () {
      if (keywords.length !== 0) {
        if (hasNoNonTextChildren(this)) {
          const res = matchKeywords($, this, keywords, mode, matchTextOnly)
          if (res !== null && res !== '') {
            resp += res + '\n\n'
          }
        }
      } else {
        resp += '<span '
        for (const attr of this.attributes) {
          resp += ` ${attr.name}="${attr.value}"`
        }
        resp += '>' + '\n'
        for (const c of this.children) {
          resp += $.html(c) + '\n'
        }
        resp += '</span>' + '\n\n'
      }
    })
  }

  return resp
}

function hasNoNonTextChildren (elem: cheerio.Element): boolean {
  for (const child of elem.children) {
    if (child.nodeType !== 3) {
      return false
    }
  }
  return true
}

const findShadowRoot = (): string => {
  const queue: Element[] = [document.body]
  let html = ''

  while (queue.length > 0) {
    const elem = queue.shift()

    if ((elem?.shadowRoot) !== null) {
      html += elem?.shadowRoot.innerHTML // Shadow root found
    }

    // Add all child elements to the queue for BFS
    if (elem?.children != null) {
      queue.push(...Array.from(elem?.children))
    }
  }

  return html
}

async function getPageHTML (page: Page): Promise<string> {
  let html = await page.content()

  // Do a breadth-first search to look for a shadow DOM
  const shadowRootHTML = await page.evaluate(findShadowRoot)
  if (shadowRootHTML !== undefined) {
    html += shadowRootHTML
  }

  // Check for shadow DOM in iframes
  for (const f of page.frames()) {
    try {
      const shadowRootHTML = await f.locator('body').evaluate(findShadowRoot, { timeout: 7000 })
      if (shadowRootHTML !== undefined) {
        html += shadowRootHTML
      }
    } catch (e) {
    }
  }

  return html
}

function getActionInstructions (action: string, args: Record<string, string>): string {
  switch (action) {
    case 'fill':
      return `You are an expert with deep knowledge of web pages and HTML elements.
      Based on the provided HTML below, return the CSS selector that can be used to locate the element described by the user input.
      The element should be a text input, textarea, or contenteditable.
      Pseudo-classes like :has-text() (i.e. article:has-text("hello")) are available for use.
      Do not escape the selector unless necessary.
      Return exactly one selector that is the best match, and don't quote the output.
      If possible, use a selector that is specific to the keywords that were provided.

      User Input: ${args.userInput}

      HTML:
      ${args.elementData}`

    case 'select':
      return `You are an expert with deep knowledge of web pages and HTML elements.
      Based on the provided HTML below, return the CSS selector that can be used to locate the select element described by the user input.
      Pseudo-classes like :has-text() (i.e. article:has-text("hello")) are available for use.
      Do not escape the selector unless necessary.
      Also determine which of the options in the select element should be selected based on the user selection.
      Return a JSON object with the keys 'selector' and 'option'.

      User Input: ${args.userInput}
      User Selection: ${args.userSelection}

      HTML:
      ${args.elementData}


      Example output: {"selector":"[data-testid='SearchForm-sortBy']","option":"newest"}`

    case 'screenshot':
      return `You are an expert with deep knowledge of web pages and HTML elements.
      Based on the provided HTML below, return the CSS selector that can be used to locate the element described by the user input.
      Pseudo-classes like :has-text() (i.e. article:has-text("hello")) are available for use.
      The user is trying to take a screenshot of part of the page, so if there are multiple elements that look relevant,
      select the outermost one.
      Do not escape the selector unless necessary.
      Return exactly one selector that is the best match, and don't quote the output.

      User Input: ${args.userInput}

      HTML:
      ${args.elementData}`

    default:
      return `You are an expert with deep knowledge of web pages and HTML elements.
      Based on the provided HTML below, return the CSS selector that can be used to locate the element described by the user input.
      Pseudo-classes like :has-text() (i.e. article:has-text("hello")) are available for use.
      Do not escape the selector unless necessary.
      Return exactly one selector that is the best match, and don't quote the output.
      If possible, use a selector that is specific to the keywords that were provided.

      User Input: ${args.userInput}

      HTML:
      ${args.elementData}`
  }
}

const findKeywordsInElement = ($: cheerio.CheerioAPI, elem: cheerio.Element, keywords: string[], matchAll: boolean): string => {
  for (const attr of elem.attributes) {
    if (keywords.every(keyword => attr.value.toLowerCase().includes(keyword.toLowerCase())) ||
      keywords.every(keyword => attr.name.toLowerCase().includes(keyword.toLowerCase()))) {
      return $.html(elem)
    }
  }

  let result = ''
  for (const c of $(elem).children()) {
    result += findKeywordsInElement($, c, keywords, matchAll)
  }
  return result
}
