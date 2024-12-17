import fetch from 'node-fetch'

export async function getAllProjects(baseUrl, auth) {
    try {
      const response = await fetch(`${baseUrl}/project`, {
        method: 'GET',
        headers: {
          'Authorization': auth,
          'Accept': 'application/json'
        }
      })
  
      const data = await response.json()
  
      console.log('Projects:')
      data.forEach(project => {
        console.log(`ID: ${project.id}, Key: ${project.key}, Name: ${project.name}`)
      })
  
      return data.map(project => ({
        id: project.id,
        key: project.key,
        name: project.name
      }))
    } catch (error) {
      console.error('Error fetching projects:', error)
      throw error
    }
  }
  
export async function getProject(baseUrl, auth, projectIdOrKey) {
    try {
      const response = await fetch(`${baseUrl}/project/${projectIdOrKey}`, {
        method: 'GET',
        headers: {
          'Authorization': auth,
          'Accept': 'application/json'
        }
      })
  
      const data = await response.json()
  
      // Extracting useful information
      console.log(`Project Details:`)
      console.log(`- ID: ${data.id}`)
      console.log(`- Key: ${data.key}`)
      console.log(`- Name: ${data.name}`)
      console.log(`- Description: ${data.description || 'N/A'}`)
      console.log(`- Project Type: ${data.projectTypeKey}`)
      console.log(`- Style: ${data.style}`)
      console.log(`- Lead: ${data.lead.displayName} (${data.lead.active ? 'Active' : 'Inactive'})`)
      console.log(`- Avatar URL: ${data.avatarUrls['48x48']}`)
      console.log(`- Issue Types:`)
      data.issueTypes.forEach(issueType => {
        console.log(`  - ${issueType.name}: ${issueType.description}`)
      })
      console.log(`- Roles:`)
      for (const [role, url] of Object.entries(data.roles)) {
        console.log(`  - ${role}: ${url}`)
      }
  
      return data
    } catch (error) {
      console.error(`Error fetching project ${projectIdOrKey}:`, error)
      throw error
    }
  }



export async function listIssues(baseUrl, auth, projectKeyOrId = null) {
  try {
    // Build the JQL query dynamically
    // TODO: add more fields to the function to support more complex queries, such as status, assignee, reporter, created, updated, etc.
    const jql = projectKeyOrId ? `project = ${projectKeyOrId}` : '' // Fetch all issues if no projectKey
    const query = jql ? `?jql=${encodeURIComponent(jql)}` : ''

    const response = await fetch(`${baseUrl}/search${query}`, {
      method: 'GET',
      headers: {
        'Authorization': auth,
        'Accept': 'application/json'
      }
    })

    const data = await response.json()

    // Extract and print a summary of each issue
    console.log(`Issues ${projectKeyOrId ? `in Project: ${projectKeyOrId}` : 'across all projects'}`)
    data.issues.forEach(issue => {
      console.log(`- Key: ${issue.key}, Summary: ${issue.fields.summary}, Status: ${issue.fields.status.name}, Assignee: ${issue.fields.assignee ? issue.fields.assignee.displayName : 'Unassigned'}, Reporter: ${issue.fields.reporter.displayName}, Created: ${issue.fields.created}, Created At: ${issue.fields.createdAt}`)
    })

    // Return a filtered list of issues
    return data.issues.map(issue => ({
      id: issue.id,
      key: issue.key,
      summary: issue.fields.summary,
      status: issue.fields.status.name,
      assignee: issue.fields.assignee ? issue.fields.assignee.displayName : 'Unassigned',
      created: issue.fields.created
    }))
  } catch (error) {
    console.error(`Error fetching issues ${projectKeyOrId ? `for project ${projectKeyOrId}` : ''}:`, error)
    throw error
  }
}



export async function getIssue(baseUrl, auth, issueIdOrKey) {
  try {
    const response = await fetch(`${baseUrl}/issue/${issueIdOrKey}`, {
      method: 'GET',
      headers: {
        'Authorization': auth,
        'Accept': 'application/json'
      }
    })

    const data = await response.json()

    // Extract meaningful fields
    console.log('Issue Details:')
    console.log(`- ID: ${data.id}`)
    console.log(`- Key: ${data.key}`)
    console.log(`- Summary: ${data.fields.summary}`)
    console.log(`- Status: ${data.fields.status.name}`)
    console.log(`- Assignee: ${data.fields.assignee ? data.fields.assignee.displayName : 'Unassigned'}`)
    console.log(`- Reporter: ${data.fields.reporter.displayName}`)
    console.log(`- Created At: ${data.fields.created}`)
    console.log()

    return {
      id: data.id,
      key: data.key,
      summary: data.fields.summary,
      status: data.fields.status.name,
      assignee: data.fields.assignee ? data.fields.assignee.displayName : 'Unassigned',
      reporter: data.fields.reporter.displayName,
      created: data.fields.created
    }
  } catch (error) {
    console.error(`Error fetching issue ${issueIdOrKey}:`, error)
    throw error
  }
}


export async function getIssueTypes(baseUrl, auth, projectIdOrKey, isFunctionCall = true) {
  try {
    const response = await fetch(`${baseUrl}/issue/createmeta/${projectIdOrKey}/issuetypes`, {
      method: 'GET',
      headers: {
        'Authorization': auth,
        'Accept': 'application/json'
      }
    })

    const data = await response.json()
    if (isFunctionCall) {
      console.log(`Issue types for project ${projectIdOrKey}: ${JSON.stringify(data, null, 2)}`)
    }

    return data
  } catch (error) {
    if (isFunctionCall) {
      console.error(`Error fetching issue types for project ${projectIdOrKey}:`, error)
    }
    throw error
  }
}

export async function createIssue(
  baseUrl,
  auth,
  projectId,
  summary,
  description = '',
  issueTypeId = '', // different type of project can have different issue types
  priority = '', 
  assignee = '', 
  reporter = '', 
  // dueDate = '', // TODO: add dueDate
  // parentKey = '', // TODO: add parentKey to support sub-issues
) {
  // Build the body dynamically
  const bodyData = {
    fields: {
      project: { id: projectId },
      summary: summary,
    }
  }

  // Add optional fields only if they are not empty
  if (issueTypeId) {
    bodyData.fields.issuetype = { id: issueTypeId }
  }
  else { // An issueType is required, so we get the first one for this project
    const issueTypeList = await getIssueTypes(baseUrl, auth, projectId, false)
    issueTypeId = issueTypeList.issueTypes[0].id
    bodyData.fields.issuetype = { id: issueTypeId }
  }

  if (priority) {
    bodyData.fields.priority = { id: priority }
  }

  if (assignee) {bodyData.fields.assignee = { id: assignee }}
  if (reporter) {bodyData.fields.reporter = { id: reporter }}
  if (description) {
    bodyData.fields.description = {
      content: [
        {
          content: [{ text: description, type: 'text' }],
          type: 'paragraph'
        }
      ],
      type: 'doc',
      version: 1
    }
  }
  // if (dueDate) bodyData.fields.duedate = dueDate
  // if (parentKey) bodyData.fields.parent = { key: parentKey }

  try {
    const response = await fetch(`${baseUrl}/issue`, {
      method: 'POST',
      headers: {
        'Authorization': auth,
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(bodyData)
    })

    const data = await response.json()
    if (response.status >= 200 && response.status < 300) {
      console.log('Issue Created:', JSON.stringify(data, null, 2))
    } else {
      console.log('Issue Creation Failed, Error:', JSON.stringify(data, null, 2))
    }
    return data
  } catch (error) {
    console.error('Error creating issue:', error)
    throw error
  }
}



  
  

