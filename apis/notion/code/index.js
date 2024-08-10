import {APIResponseError, Client} from "@notionhq/client"
import {createPage, printPageProperties, recursivePrintChildBlocks} from "./src/pages.js"
import {printDatabaseRow} from "./src/database.js";
import {printSearchResults} from "./src/search.js";

if (process.argv.length !== 3) {
    console.error('Usage: node index.js <command>')
    process.exit(1)
}

const command = process.argv[2]
const token = process.env.NOTION_TOKEN

const notion = new Client({auth: token})

async function main() {
    try {
        switch (command) {
            case "search":
                printSearchResults(await notion.search({query: process.env.QUERY}))
                break
            case "getPage":
                await printPageProperties(notion, process.env.ID)
                console.log("Page Contents:")
                await recursivePrintChildBlocks(notion, process.env.ID)
                break
            case "getDatabase":
                const response = await notion.databases.query({database_id: process.env.ID})
                for (const row of response.results) {
                    if (row.object === "page") {
                        await printDatabaseRow(notion, row)
                    }
                }
                break
            case "createPage":
                await createPage(notion, process.env.NAME, process.env.CONTENTS, process.env.PARENTPAGEID)
                break
            default:
                console.log(`Unknown command: ${command}`)
                process.exit(1)
        }
    } catch (error) {
        // We use console.log instead of console.error here so that it goes to stdout
        if (error instanceof APIResponseError) {
            console.log(error.message)
        } else {
            console.log("Got an unknown error")
        }
    }
}

await main()
