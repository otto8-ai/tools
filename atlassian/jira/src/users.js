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
        const allUsers = []
        for (const user of data) {
            if (includeAppUsers || user.accountType != "app") {
                allUsers.push({
                    accountType: user.accountType,
                    accountId: user.accountId,
                    displayName: user.displayName,
                    emailAddress: user.emailAddress,
                    active: user.active,
                    locale: user.locale
                })
            }
        }
        console.log(JSON.stringify({allUsers}))
    } catch (error) {
        console.log(JSON.stringify({error}))
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
        const user = {
            accountType: data.accountType,
            accountId: data.accountId,
            displayName: data.displayName,
            emailAddress: data.emailAddress,
            active: data.active,
            locale: data.locale
        }
        console.log(JSON.stringify({user}))
    } catch (error) {
        console.log(JSON.stringify({error}))
    }
}

export async function getCurrentUser(baseUrl, auth) {
    try {
        const response = await fetch(`${baseUrl}/myself`, {
            method: 'GET',
            headers: {
                'Authorization': auth,
                'Accept': 'application/json'
            }
        })
        const currentUser = await response.json()
        console.log(JSON.stringify({currentUser}))
    } catch (error) {
        console.log(JSON.stringify({error}))
    }
}
