import {getDashboard, getAllDashboard, getDashboardGadgets } from './src/dashboards.js'
import { Version3Client } from 'jira.js'

const token = process.env.JIRA_TOKEN

// TODO: once Oauth2.0 is implemented, this URL will be something like https://api.atlassian.com/ex/jira/<cloudId>/rest/api/3
// where the cloudId can be obtained fron Oauth. see more details here https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/#implementing-oauth-2-0--3lo-
const baseUrl = 'https://acorn-team-yb.atlassian.net/rest/api/3'

// TODO: replace auth with Oauth2.0 token. it will look like Bearer aBCxYz654123
const auth = `Basic ${Buffer.from(
            `yingbei@acorn.io:${token}`
          ).toString('base64')}`


getDashboard(baseUrl, auth, "10001")

// getDashboardGadgets(baseUrl, auth, "10001")

