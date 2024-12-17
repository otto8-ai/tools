import { listJiraSites } from './src/sites.js'
import { getAllProjects, getProject , getIssue, listIssues, createIssue, getIssueTypes} from './src/issues.js'
import { getPrioritySchemes, getAvailablePriorities } from './src/priority.js'
import { getAllUsers, getUser } from './src/users.js'

const TOKEN = process.env.ATLASSIAN_OAUTH_TOKEN

if (process.argv.length !== 3) {
    console.error('Usage: node index.js <command>')
    process.exit(1)
}

const command = process.argv[2]

async function main() {
    let cloudId = ""
    const auth = `Bearer ${TOKEN}`
    try {
        // get the cloudId corresponding to the access token from the Jira instance
        const response = await fetch('https://api.atlassian.com/oauth/token/accessible-resources', {
            method: 'GET',
            headers: {
              'Authorization': auth,
              'Accept': 'application/json',
            },
          })

        if (!response.ok) {
            throw new Error(`Error: ${response.status} ${response.statusText}`)
        }

        const resources = await response.json()
        cloudId = resources[0].id
        console.log(JSON.stringify({accessibleResources: resources}))
    } catch (error) {
        console.error("Failed to get Jira cloudId:", error.message)
        throw error
    }


    const baseUrl = `https://api.atlassian.com/ex/jira/${cloudId}/rest/api/3`
    try {
        switch (command) {
            case "listJiraSites":
                const sites = await listJiraSites(auth)
                console.log(JSON.stringify({jiraSites: sites}))
                break
            case "createIssue":
                await createIssue(
                    baseUrl,
                    auth,
                    process.env.PROJECT_ID,
                    process.env.SUMMARY,
                    process.env.DESCRIPTION,
                    process.env.ISSUE_TYPE,
                    process.env.PRIORITY,
                    process.env.ASSIGNEE_ID,
                    process.env.REPORTER_ID
                )
                break
            case "listIssues":
                await listIssues(
                    baseUrl,
                    auth,
                    process.env.PROJECT_KEY_OR_ID
                )
                break
            case "getIssue":
                await getIssue(
                    baseUrl,
                    auth,
                    process.env.ISSUE_KEY
                )
                break
            case "getIssueTypes":
                await getIssueTypes(
                    baseUrl,
                    auth,
                    process.env.PROJECT_KEY_OR_ID
                )
                break
            case "getAvailablePriorities":
                await getAvailablePriorities(
                    baseUrl,
                    auth,
                    process.env.SCHEME_ID
                )
                break
            case "getPrioritySchemes":
                await getPrioritySchemes(
                    baseUrl,
                    auth,
                    process.env.SCHEME_ID
                )
                break
            case "getAllUsers":
                await getAllUsers(
                    baseUrl,
                    auth,
                    process.env.INCLUDE_APP_USERS
                )
                break
            case "getUser":
                await getUser(
                    baseUrl,
                    auth,
                    process.env.ACCOUNT_ID
                )
                break
            case "getCurrentUser":
                await getCurrentUser(
                    baseUrl,
                    auth
                )
                break
            case "getAllProjects":
                await getAllProjects(
                    baseUrl,
                    auth,
                    process.env.INCLUDE_APP_USERS
                )
                break
            case "getProject":
                await getProject(
                    baseUrl,
                    auth,
                    process.env.PROJECT_KEY_OR_ID
                )
                break
            default:
                console.log(`Unknown command: ${command}`)
                process.exit(1)
        }
    } catch (error) {
        // We use console.log instead of console.error here so that it goes to stdout
        console.log("Got an error:", error.message)
        throw error
    }
}

await main()
