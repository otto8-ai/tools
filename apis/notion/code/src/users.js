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

    for (let i = 0; i < max && i < users.length; i++) {
        console.log(`${users[i].name} (ID: ${users[i].id})`)
    }
}
