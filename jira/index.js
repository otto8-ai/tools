import {getDashboard, getAllDashboards, getDashboardGadgets } from './src/dashboards.js'
import { getAllProjects, getProject , getIssue, listIssues, createIssue, getIssueTypes} from './src/issues.js'
import { getPrioritySchemes, getAvailablePriorities } from './src/priority.js'
import { getAllUsers, getUser } from './src/users.js'

const token = process.env.JIRA_TOKEN

// TODO: once Oauth2.0 is implemented, this URL will be something like https://api.atlassian.com/ex/jira/<cloudId>/rest/api/3
// where the cloudId can be obtained fron Oauth. see more details here https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/#implementing-oauth-2-0--3lo-
const baseUrl = 'https://acorn-team-yb.atlassian.net/rest/api/3'

// TODO: replace auth with Oauth2.0 token. it will look like Bearer aBCxYz654123
const auth = `Basic ${Buffer.from(
            `yingbei@acorn.io:${token}`
          ).toString('base64')}`

if (process.argv.length !== 3) {
    console.error('Usage: node index.js <command>')
    process.exit(1)
}

const command = process.argv[2]

async function main() {
    try {
        switch (command) {
            case "createIssue":
                await createIssue(baseUrl, auth, process.env.PROJECTID, process.env.SUMMARY, process.env.DESCRIPTION, process.env.ISSUETYPE, process.env.PRIORITY, process.env.ASSIGNEEID, process.env.REPORTERID)
                break
            case "listIssues":
                await listIssues(baseUrl, auth, process.env.PROJECTKEYORID)
                break
            case "getIssue":
                await getIssue(baseUrl, auth, process.env.ISSUEKEY)
                break
            case "getIssueTypes":
                await getIssueTypes(baseUrl, auth, process.env.PROJECTKEYORID)
                break
            case "getAvailablePriorities":
                await getAvailablePriorities(baseUrl, auth,  process.env.SCHEMEID)
                break
            case "getPrioritySchemes":
                await getPrioritySchemes(baseUrl, auth)
                break
            case "getAllUsers":
                await getAllUsers(baseUrl, auth, process.env.INCLUDEAPPUSERS)
                break
            case "getUser":
                await getUser(baseUrl, auth, process.env.ACCOUNTID)
                break
            case "getAllProjects":
                await getAllProjects(baseUrl, auth)
                break
            case "getProject":
                await getProject(baseUrl, auth, process.env.PROJECTKEYORID)
                break
            case "getAllDashboards":
                await getAllDashboards(baseUrl, auth)
                break
            case "getDashboard":
                await getDashboard(baseUrl, auth, process.env.DASHBOARDID)
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
