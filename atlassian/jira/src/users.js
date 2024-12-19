export async function listUsers(client, includeAppUsers = false) {
    try {
        const { data: allUsers } = await client.get('/users')

        const users = [] 
        for (const user of allUsers) {
            if (includeAppUsers || user.accountType != "app") {
                users.push({
                    accountType: user.accountType,
                    accountId: user.accountId,
                    displayName: user.displayName,
                    emailAddress: user.emailAddress,
                    active: user.active,
                    locale: user.locale
                })
            }
        }

        return users
    } catch (error) {
        throw new Error(`Error fetching users: ${error.message}`);
    }
}

export async function getUser(client, accountId) {
    try {
        const { data: user } = await client.get(`/user?accountId=${accountId}`)

        return {
            accountType: user.accountType,
            accountId: user.accountId,
            displayName: user.displayName,
            emailAddress: user.emailAddress,
            active: user.active,
            locale: user.locale
        }
    } catch (error) {
        throw new Error(`Error fetching user ${accountId}: ${error.message}`);
    }
}

export async function getCurrentUser(client) {
    try {
        const { data: currentUser } = await client.get('/myself')
        return currentUser
    } catch (error) {
        throw new Error(`Error fetching current user: ${error.message}`);
    }
}
