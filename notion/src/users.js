import {GPTScript} from "@gptscript-ai/gptscript";

export async function listUsers(client, max) {
    if (max === undefined) {
        max = 999999999 // basically unlimited
    }
    let nextCursor = undefined
    let users = []
    while (true) {
        let response
        if (nextCursor === undefined) {
            response = await client.users.list()
        } else {
            response = await client.users.list({start_cursor: nextCursor})
        }
        users = users.concat(response.results.map(r => { return {id: r.id, name: r.name}}))

        if (response.has_more === false || users.length >= max) {
            break
        }
        nextCursor = response.next_cursor
    }

    try {
        const gptscriptClient = new GPTScript()
        let elements = users.map(user => {
            return {
                name: `${user.name} ${user.id}`,
                description: `${user.name} (ID: ${user.id})`,
                contents: `${user.name} (ID: ${user.id})`,
            }
        })

        if (max < elements.length) {
            elements = elements.slice(0, max)
        }

        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: "notion_users",
        })
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} users`)
    } catch (e) {
        console.log("Failed to create dataset:", e)
    }
}
