import fetch from 'node-fetch'

export async function getPrioritySchemes(baseUrl, auth, isFunctionCall = true) {
    try {
        const response = await fetch(`${baseUrl}/priorityscheme`, {
        method: 'GET',
        headers: {
            'Authorization': auth,
            'Accept': 'application/json'
        }
        })

        const data = await response.json()
        if (isFunctionCall) {
            console.log('Priority Schemes:', JSON.stringify(data, null, 2))
        }

        return data
    } catch (error) {
        if (isFunctionCall) {
            console.error('Error fetching priority schemes:', error)
        }
        throw error
    }
}

export async function getAvailablePriorities(baseUrl, auth, schemeId="", isFunctionCall = true) {
    try {
        if (schemeId === "") { // get Default Priority Scheme Id
            const schemeIdList = await getPrioritySchemes(baseUrl, auth, false)
            schemeId = schemeIdList.values[0].id
            console.log(`Using Default Priority Scheme Id: ${schemeId}`)
        }
        const response = await fetch(`${baseUrl}/priorityscheme/${schemeId}/priorities`, {
        method: 'GET',
        headers: {
            'Authorization': auth,
            'Accept': 'application/json'
        }
        })

        const data = await response.json()
        if (isFunctionCall) {
            console.log(`Priority Details for Scheme ${schemeId}:`, JSON.stringify(data, null, 2))
        }

        return data
    } catch (error) {
        if (isFunctionCall) {
            console.error(`Error fetching priorities for scheme ${schemeId}:`, error)
        }
        throw error
    }
}
  
  