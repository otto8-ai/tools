import {getDashboard, getAllDashboards, getDashboardGadgets } from './src/dashboards.js'
import { getAllProjects, getProject , getIssue, listIssues, createIssue, getIssueTypes} from './src/issues.js'
import { getPrioritySchemes, getAvailablePriorities } from './src/priority.js'
import { getAllUsers, getUser } from './src/users.js'

const token = process.env.JIRA_TOKEN

if (process.argv.length !== 3) {
    console.error('Usage: node index.js <command>')
    process.exit(1)
}

const command = process.argv[2]

async function main() {
    let cloudId = ""
    const auth = `Bearer ${token}`
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
        cloudId = resources[0].id // TODO: if the user has multiple Jira instances, we need to add a way to select the correct one
        
    } catch (error) {
        console.error("Failed to get Jira cloudId:", error.message)
        throw error
    }
    const baseUrl = `https://api.atlassian.com/ex/jira/${cloudId}/rest/api/3`
    
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
