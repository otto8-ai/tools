import {APIResponseError, Client} from "@notionhq/client"
import {createPage, updatePage, printPageProperties, recursivePrintChildBlocks} from "./src/pages.js"
import {addDatabaseRow, describeProperty, printDatabaseRow, updateDatabaseRow} from "./src/database.js";
import {search} from "./src/search.js";
import {listUsers} from "./src/users.js";

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
            case "listUsers":
                await listUsers(notion, process.env.MAX)
                break
            case "search":
                await search(notion, process.env.QUERY, process.env.MAX)
                break
            case "getPage":
                await printPageProperties(notion, process.env.ID)
                console.log("Page Contents:")
                await recursivePrintChildBlocks(notion, process.env.ID)
                break
            case "addDatabaseRow":
                await addDatabaseRow(notion, process.env.ID, JSON.parse(process.env.PROPERTIES))
                break
            case "updateDatabaseRow":
                await updateDatabaseRow(notion, process.env.ID, process.env.ROWID, JSON.parse(process.env.PROPERTIES))
                break
            case "getDatabaseProperties":
                const retrieval = await notion.databases.retrieve({database_id: process.env.ID})
                console.log(`Properties for database ${retrieval.title[0].plain_text}:`)
                for (const [name, property] of Object.entries(retrieval.properties)) {
                    const description = describeProperty(name, property)
                    if (description !== "") {
                        console.log(description)
                    }
                }
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
            case "updatePage":
                await updatePage(notion, process.env.PAGEID, process.env.CONTENTS, process.env.UPDATEMODE)
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
            console.log("Got an error:", error.message)
            throw error
        }
    }
}

await main()
