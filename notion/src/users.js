import {GPTScript} from "@gptscript-ai/gptscript";
import {min} from "./util.js";

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

    if (min(users.length, max) > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_DIR, "notion_users", "list of notion users")
            for (let i = 0; i < min(users.length, max); i++) {
                await gptscriptClient.addDatasetElement(process.env.GPTSCRIPT_WORKSPACE_DIR, dataset.id, users[i].name + users[i].id, "", `${users[i].name} (ID: ${users[i].id})`)
            }
            console.log(`Created dataset with ID ${dataset.id} with ${min(users.length, max)} users`)
            return
        } catch (e) {
            console.log("Error initializing GPTScript client: ", e)
            process.exit(1)
        }
    }

    for (let i = 0; i < max && i < users.length; i++) {
        console.log(`${users[i].name} (ID: ${users[i].id})`)
    }
}
