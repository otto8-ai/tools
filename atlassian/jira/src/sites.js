import fetch from 'node-fetch'

const REQUIRED_SCOPES = ['write:jira-work', 'read:jira-work', 'read:jira-user']
const hasRequiredScopes = (scopes) =>
  REQUIRED_SCOPES.every(item => scopes.includes(item));

export async function listJiraSites(auth) {
    try {
        const response = await fetch(`${baseUrl}/users`, {
            method: 'GET',
            headers: {
                'Authorization': auth,
                'Accept': 'application/json'
            }
        })

        const data = await response.json()
        const jiraSites = []
        for (const site of data) {
            if (hasRequiredScopes(site.scopes)) {
                jiraSites.push((( { scopes, ...s} = site) => s)())
            }
        }
        return {jiraSites}
    } catch (error) {
        throw {listJiraSitesError: error}
    }
}
