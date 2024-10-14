import { Client } from "@notionhq/client";
import { BlockObjectResponse } from "@notionhq/client/build/src/api-endpoints";


export async function getPageContent(client: Client, id: string, indentation = 0): Promise<string> {
  const blocks = await client.blocks.children.list({block_id: id})
  let result: string = '';
  for (let b of blocks.results) {
    let block = b as BlockObjectResponse;
    // Tables are complicated, so we handle them completely separately
    if (block.type === "table") {
      result += await printTable(client, b)
      continue
    }

    result += await printBlock(client, b as BlockObjectResponse, indentation)
    if (block.has_children && block.type !== "child_page" && block.type !== "synced_block") {
      result += await getPageContent(client, b.id, indentation + 2)
    }
  }
  return result
}

async function printBlock(client: Client, b: BlockObjectResponse, indentation: number): Promise<string> {
  let result: string = ""
  if (indentation > 0) {
    result += " ".repeat(indentation)
  }
  switch (b.type) {
    case "bookmark":
      if (b.bookmark.caption !== null && richTextArrayToString(b.bookmark.caption) !== "") {
        result += `Bookmark: ${b.bookmark.url} (${richTextArrayToString(b.bookmark.caption)})`
      } else {
        result += `Bookmark: ${b.bookmark.url}`
      }
      break
    case "bulleted_list_item":
      result += `- ${richTextArrayToString(b.bulleted_list_item.rich_text)}`
      break
    case "callout":
      result += `> ${richTextArrayToString(b.callout.rich_text)}`
      break
    case "child_database":
      result += `Child Database: ${b.child_database.title}`
      break
    case "child_page":
      result += `Child Page: ${b.child_page.title}`
      break
    case "code":
      if (b.code.language !== null) {
        result += "```" + b.code.language + "\n"
      } else {
        result += "```\n"
      }
      result += richTextArrayToString(b.code.rich_text)
      result += "\n```"
      if (b.code.caption !== null && richTextArrayToString(b.code.caption) !== "") {
        result += `\n(${richTextArrayToString(b.code.caption)})`
      }
      break
    case "divider":
      result += "-------------------------------------"
      break
    case "embed":
      result += `Embed: ${b.embed.url}`
      break
    case "equation":
      result += `Equation: ${b.equation.expression}`
      break
    case "file":
      result += fileToString("File", b.file)
      break
    case "heading_1":
      result += `# ${richTextArrayToString(b.heading_1.rich_text)}`
      break
    case "heading_2":
      result += `## ${richTextArrayToString(b.heading_2.rich_text)}`
      break
    case "heading_3":
      result += `### ${richTextArrayToString(b.heading_3.rich_text)}`
      break
    case "image":
      result += fileToString("Image", b.image)
      break
    case "link_preview":
      result += b.link_preview.url
      break
    case "numbered_list_item":
      result += `1. ${richTextArrayToString(b.numbered_list_item.rich_text)}`
      break
    case "paragraph":
      result += richTextArrayToString(b.paragraph.rich_text)
      break
    case "pdf":
      result += fileToString("PDF", b.pdf)
      break
    case "quote":
      result += "\"\"\"\n"
      result += richTextArrayToString(b.quote.rich_text)
      result += "\n\"\"\""
      break
    case "synced_block":
      if (b.synced_block.synced_from !== null) {
        await getPageContent(client, b.synced_block.synced_from.block_id, indentation)
      }
      break
    case "to_do":
      if (b.to_do.checked) {
        result += `[x] ${richTextArrayToString(b.to_do.rich_text)}`
      } else {
        result += `[ ] ${richTextArrayToString(b.to_do.rich_text)}`
      }
      break
    case "toggle":
      result += `> ${richTextArrayToString(b.toggle.rich_text)}`
      break
    case "video":
      result += fileToString("Video", b.video)
      break
  }
  return result.replace("\n", "\n" + " ".repeat(indentation))
}

export function richTextArrayToString(richTextArray: any[]) {
  let result = ""
  for (let r of richTextArray) {
    result += r.plain_text + " "
  }
  return result
}

function fileToString(prefix: any, file: any) {
  let result = ""
  if (file.type === "file") {
    result = `${prefix}: ${file.file.url} (expires ${file.file.expiry_time})`
  } else if (file.type === "external") {
    result = `External ${prefix}: ${file.external.url}`
  }
  if (file.caption !== null && richTextArrayToString(file.caption) !== "") {
    result += ` (${richTextArrayToString(file.caption)})`
  }
  return result
}

async function printTable(client: Client, table: any): Promise<string> {
  let result = ""
  const children = await client.blocks.children.list({block_id: table.id})
  if (table.table.has_column_header && children.results.length > 0) {
    result += printTableRow((children.results[0] as any).table_row, table.table.has_row_header, true)
    for (let i = 1; i < children.results.length; i++) {
      result += printTableRow((children.results[i] as any).table_row, table.table.has_row_header, false)
    }
  } else {
    for (let r of children.results) {
      result += printTableRow((r as any).table_row, table.table.has_row_header, false)
    }
  }
  return result;
}

function printTableRow(row: any, boldFirst: any, boldAll: any): string {
  let result = "|"
  if (boldAll) {
    for (let c of row.cells) {
      result += ` **${richTextArrayToString(c)}** |`
    }
    let len = result.length
    result += "\n|" + "-".repeat(len - 2) + "|"
  } else if (boldFirst && row.cells.length > 0) {
    result += ` **${richTextArrayToString(row.cells[0])}** |`
    for (let i = 1; i < row.cells.length; i++) {
      result += ` ${richTextArrayToString(row.cells[i])} |`
    }
  } else {
    for (let c of row.cells) {
      result += ` ${richTextArrayToString(c)} |`
    }
  }
  return result
}
