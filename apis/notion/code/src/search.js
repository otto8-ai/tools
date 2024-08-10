export function printSearchResults(res) {
    let pages = []
    let databases = []
    for (let r of res.results) {
        switch (r.object) {
            case "page":
                pages.push(r)
                break
            case "database":
                databases.push(r)
                break
        }
    }
    if (pages.length > 0) {
        printPages(pages)
    }
    console.log("")
    if (databases.length > 0) {
        printDatabases(databases)
    }
}

function printPages(pages) {
    console.log("Pages:")
    for (let page of pages) {
        console.log(`- ID: ${page.id}`)
        if (page.properties.title !== undefined && page.properties.title.title.length > 0) {
            console.log(`  Title: ${page.properties.title.title[0].plain_text}`)
        } else if (page.properties.Name !== undefined && page.properties.Name.title.length > 0) {
            console.log(`  Name: ${page.properties.Name.title[0].plain_text}`)
        }
        console.log(`  URL: ${page.url}`)
        console.log(`  Parent Type: ${page.parent.type}`)
        if (page.parent.type === "database_id") {
            console.log(`  Parent Database ID: ${page.parent.database_id}`)
        } else if (page.parent.type === "page_id") {
            console.log(`  Parent Page ID: ${page.parent.page_id}`)
        } else if (page.parent.type === "block_id") {
            console.log(`  Parent Block ID: ${page.parent.block_id}`)
        }
    }
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
