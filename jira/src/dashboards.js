import fetch from 'node-fetch'

export async function getAllDashboards(baseUrl, auth) {
    try {
        const response = await fetch(`${baseUrl}/dashboard`, {
            method: 'GET',
            headers: {
            'Authorization': auth,
            'Accept': 'application/json'
            }
        })
    
        const data = await response.json()
        console.log('Dashboards:')
        data.dashboards.forEach(dashboard => {
        const domain = new URL(dashboard.self).origin
        console.log(`ID: ${dashboard.id}, Name: ${dashboard.name}, view-URL: ${domain}${dashboard.view}`)
        })
  
      return data
    } catch (error) {
        console.error(error)
        throw error
    }
}

export async function getDashboard(baseUrl, auth, id) {
    try {
        const response = await fetch(`${baseUrl}/dashboard/${id}`, {
        method: 'GET',
        headers: {
            'Authorization': auth,
            'Accept': 'application/json'
        }
        })

        const data = await response.json()
        const domain = new URL(data.self).origin

        console.log(`Got dashboard. ID: ${id}, Dashboard view-URL: ${domain}${data.view}, Content:\n${JSON.stringify(data, null, 2)}`)

        return data
    } catch (error) {
        console.error(error)
        throw error
    }
}


export async function getDashboardGadgets(baseUrl, auth, dashboardId) {
    try {
        const response = await fetch(`${baseUrl}/dashboard/${dashboardId}/gadget`, {
        method: 'GET',
        headers: {
            'Authorization': auth,
            'Accept': 'application/json'
        }
        })

        const data = await response.json()
        console.log(`Got gadgets for dashboard ${dashboardId}: ${JSON.stringify(data, null, 2)}`)

        return data
    } catch (error) {
        console.error(`Error fetching gadgets for dashboard ${dashboardId}:`, error)
        throw error
    }
}

  

  


  