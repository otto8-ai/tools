import fetch from 'node-fetch'

export async function getAllUsers(baseUrl, auth, includeAppUsers = false) {
    try {
        const response = await fetch(`${baseUrl}/users`, {
            method: 'GET',
            headers: {
            'Authorization': auth,
            'Accept': 'application/json'
        }
        })

        const data = await response.json()
        const users = []
        for (const user of data) {
            if (includeAppUsers || user.accountType != "app") {
                const userData = {accountType: user.accountType, accountId: user.accountId, displayName: user.displayName, emailAddress: user.emailAddress, active: user.active, locale: user.locale}
                users.push(userData)
            }
        }
        console.log(`All users: ${JSON.stringify(users, null, 2)}`)
        return users
    } catch (error) {
        console.error(`Error getting all users: ${error}`)
        return null
    }
}

export async function getUser(baseUrl, auth, accountId) {
    try {
        const response = await fetch(`${baseUrl}/user?accountId=${accountId}`, {
          method: 'GET',
          headers: {
            'Authorization': auth,
            'Accept': 'application/json'
          }
        })  

        const data = await response.json()
        const user = {accountType: data.accountType, accountId: data.accountId, displayName: data.displayName, emailAddress: data.emailAddress, active: data.active, locale: data.locale}
        console.log(`User: ${JSON.stringify(user, null, 2)}`)
        return user
    } catch (error) {
        console.error(`Error getting user ${accountId}: ${error}`)
        return null
    }
}