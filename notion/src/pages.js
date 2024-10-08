import {getPropertyString} from "./database.js"

export async function createPage(client, name, contents, parentPageId) {
    const page = await client.pages.create({
        parent: {
            page_id: parentPageId,
        },
        properties: {
            title: [{type: "text", text: {content: name}}],
        },
        children: [{
            paragraph: {
                rich_text: [{type: "text", text: {content: contents}}],
            }
        }]
    })
    console.log(`Created page with ID: ${page.id}`)
}

export async function printPageProperties(client, id) {
    const page = await client.pages.retrieve({page_id: id})
    console.log("Page Properties:")
    for (const [propertyName, property] of Object.entries(page.properties)) {
        console.log(`${propertyName}: ${await getPropertyString(client, property)}`)
    }
    console.log("")
}

export async function recursivePrintChildBlocks(client, id, indentation = 0) {
    const blocks = await client.blocks.children.list({block_id: id})
    for (let b of blocks.results) {
        // Tables are complicated, so we handle them completely separately
        if (b.type === "table") {
            await printTable(client, b)
            continue
        }

        await printBlock(client, b, indentation)
        if (b.has_children && b.type !== "child_page" && b.type !== "synced_block") {
            await recursivePrintChildBlocks(client, b.id, indentation + 2)
        }
    }
}

async function printBlock(client, b, indentation) {
    let result = ""
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
        case "database":
            // TODO - test this one out
            //   Is there even a way to get a database block?
            result += `Mentioned Database ID: ${b.database.id}`
            break
        case "date":
            if (b.date.end !== null) {
                result += `${b.date.start} - ${b.date.end}`
            } else {
                result += b.date.start
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
        case "page":
            result += `Mentioned Page ID: ${b.page.id}`
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
                await recursivePrintChildBlocks(client, b.synced_block.synced_from.block_id, indentation)
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
        case "user":
            if (b.user.name !== null && b.user.name !== "") {
                result += `Mentioned User: ${b.user.name} (ID: ${b.user.id})`
            } else {
                result += `Mentioned User ID: ${b.user.id}`
            }
            break
        case "video":
            result += fileToString("Video", b.video)
            break
    }
    result = result.replaceAll("\n", "\n" + " ".repeat(indentation))
    console.log(result)
}

export function richTextArrayToString(richTextArray) {
    let result = ""
    for (let r of richTextArray) {
        result += r.plain_text + " "
    }
    return result
}

function fileToString(prefix, file) {
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

async function printTable(client, table) {
    const children = await client.blocks.children.list({block_id: table.id})
    if (table.table.has_column_header && children.results.length > 0) {
        printTableRow(children.results[0].table_row, table.table.has_row_header, true)
        for (let i = 1; i < children.results.length; i++) {
            printTableRow(children.results[i].table_row, table.table.has_row_header, false)
        }
    } else {
        for (let r of children.results) {
            printTableRow(r.table_row, table.table.has_row_header, false)
        }
    }
}

function printTableRow(row, boldFirst, boldAll) {
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
    console.log(result)
}
