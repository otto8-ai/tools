import {GPTScript} from "@gptscript-ai/gptscript";

export async function search(client, query, max) {
    if (max === undefined) {
        max = 999999999 // basically unlimited
    }
    let nextCursor = undefined
    let results = []
    while (true) {
        let response
        if (nextCursor === undefined) {
            response = await client.search({query: query})
        } else {
            response = await client.search({query: query, start_cursor: nextCursor})
        }
        results = results.concat(response.results)

        if (response.has_more === false || results.length >= max) {
            break
        }
        nextCursor = response.next_cursor
    }

    if (results.length === 0) {
        console.log("No results found")
        return
    }

    try {
        const gptscriptClient = new GPTScript()
        let elements = results.map(result => {
            let name = ""
            if (result.properties.title !== undefined && result.properties.title.title.length > 0) {
                name = result.properties.title.title[0].plain_text
            } else if (result.properties.Name !== undefined && result.properties.Name.title.length > 0) {
                name = result.properties.Name.title[0].plain_text
            } else if (result.title !== undefined && result.title.length > 0) {
                name = result.title[0].plain_text
            }

            let type = "page"
            if (result.object === "database") {
                type = "database"
            }

            return {
                name: `${result.id}`,
                description: `Notion ${type} named ${name}`,
                contents: resultToString(result, name),
            }
        })

        if (max < elements.length) {
            elements = elements.slice(0, max)
        }

        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `${query}_notion_search`,
            description: `search results from Notion for query ${query}`
        })
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} search results`)
    } catch (e) {
        console.log("Failed to create dataset:", e)
    }
}

function resultToString(res, name) {
    let str = ''
    switch (res.object) {
        case "page":
            str += `- ID: ${res.id}\n`
            str += `  Name: ${name}\n`
            str += `  URL: ${res.url}\n`
            str += `  Type: page\n`
            str += `  Parent Type: ${res.parent.type}\n`
            if (res.parent.type === "database_id") {
                str += `  Parent Database ID: ${res.parent.database_id}\n`
            } else if (res.parent.type === "page_id") {
                str += `  Parent Page ID: ${res.parent.page_id}\n`
            } else if (res.parent.type === "block_id") {
                str += `  Parent Block ID: ${res.parent.block_id}\n`
            }
            break
        case "database":
            str += `- Title: ${name}\n`
            str += `  ID: ${res.id}\n`
            str += `  URL: ${res.url}\n`
            str += `  Type: database\n`
            if (res.description.length > 0) {
                str += `  Description: ${res.description[0].plain_text}\n`
            }
            if (res.parent.type !== "") {
                str += `  Parent Type: ${res.parent.type}\n`
            }
            if (res.parent.type === "database_id") {
                str += `  Parent Database ID: ${res.parent.database_id}\n`
            } else if (res.parent.type === "page_id") {
                str += `  Parent Page ID: ${res.parent.page_id}\n`
            } else if (res.parent.type === "block_id") {
                str += `  Parent Block ID: ${res.parent.block_id}\n`
            }
            break
    }
    return str
}

function printDatabases(dbs) {
    console.log("Databases:")
    for (let db of dbs) {
        console.log(`- Title: ${db.title[0].plain_text}`)
        console.log(`  ID: ${db.id}`)
        console.log(`  URL: ${db.url}`)
        if (db.description.length > 0) {
            console.log(`  Description: ${db.description[0].plain_text}`)
        }
        if (db.parent.type !== "") {
            console.log(`  Parent Type: ${db.parent.type}`)
        }
        if (db.parent.type === "database_id") {
            console.log(`  Parent Database ID: ${db.parent.database_id}`)
        } else if (db.parent.type === "page_id") {
            console.log(`  Parent Page ID: ${db.parent.page_id}`)
        } else if (db.parent.type === "block_id") {
            console.log(`  Parent Block ID: ${db.parent.block_id}`)
        }
    }
}
