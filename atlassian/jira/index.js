import axios from 'axios'
import { listJiraSites } from './src/sites.js'
import { listProjects, getProject } from './src/projects.js'
import { searchIssues, listIssues, getIssue, createIssue, editIssue } from './src/issues.js'
import { addComment, listComments } from './src/comments.js'
import { listPriorities } from './src/priorities.js'
import { listUsers, getUser, getCurrentUser } from './src/users.js'

if (process.argv.length !== 3) {
    console.error('Usage: node index.js <command>')
    process.exit(1)
}

const command = process.argv[2]

async function main() {

    try {
        if (!process.env.ATLASSIAN_OAUTH_TOKEN) {
            throw new Error("ATLASSIAN_OAUTH_TOKEN is required")
        }

        let baseUrl
        if (command !== 'listJiraSites') {
            if (!process.env.SITE_ID) {
                throw new Error('site_id argument not provided')
            }
            baseUrl = `https://api.atlassian.com/ex/jira/${process.env.SITE_ID}/rest/api/3`
        }
        const client = axios.create({
            baseURL: baseUrl,
            headers: {
                'Authorization': `Bearer ${process.env.ATLASSIAN_OAUTH_TOKEN}`,
                'Accept': 'application/json',
            },
        })

        let result = null
        switch (command) {
            case "listJiraSites":
                result = await listJiraSites(client)
                break
            case "createIssue":
                result = await createIssue(
                    client,
                    process.env.PROJECT_ID,
                    process.env.SUMMARY,
                    process.env.DESCRIPTION,
                    process.env.ISSUE_TYPE_ID,
                    process.env.PRIORITY_ID,
                    process.env.ASSIGNEE_ID,
                    process.env.REPORTER_ID
                )
                break
            case "editIssue":
                result = await editIssue(
                    client,
                    process.env.ISSUE_ID_OR_KEY,
                    process.env.NEW_SUMMARY,
                    process.env.NEW_DESCRIPTION,
                    process.env.NEW_ASSIGNEE_ID,
                    process.env.NEW_PRIORITY_ID,
                    process.env.NEW_NAME,
                    process.env.NEW_STATUS_NAME
                )
                break
            case "searchIssues":
                result = await searchIssues(client, process.env.JQL_QUERY)
                break
            case "listIssues":
                result = await listIssues(client, process.env.PROJECT_ID_OR_KEY)
                break
            case "getIssue":
                result = await getIssue(client, process.env.ISSUE_ID_OR_KEY)
                break
            case "addComment":
                result = await addComment(client, process.env.ISSUE_ID_OR_KEY, process.env.COMMENT_BODY)
                break
            case "listComments":
                result = await listComments(client, process.env.ISSUE_ID_OR_KEY)
                break
            case "listPriorities":
                result = await listPriorities(client, process.env.SCHEME_ID)
                break
            case "listUsers":
                result = await listUsers(client, process.env.INCLUDE_APP_USERS)
                break
            case "getUser":
                result = await getUser(client, process.env.ACCOUNT_ID)
                break
            case "getCurrentUser":
                result = await getCurrentUser(client)
                break
            case "listProjects":
                result = await listProjects(client)
                break
            case "getProject":
                result = await getProject(client, process.env.PROJECT_ID_OR_KEY)
                break
            default:
                throw new Error(`Unknown command: ${command}`)
        }
        console.log(JSON.stringify(result))
    } catch (error) {
        // We use console.log instead of console.error here so that it goes to stdout
        console.log(error)
        process.exit(1)
    }
}

await main()
