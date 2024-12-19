export async function searchIssues(client, jql) {
  try {
    const query = jql ? `?jql=${encodeURIComponent(jql)}` : '';

    const { data } = await client.get(`/search${query}`)
    const issues = data.issues

    return issues.map(issue => ({
      id: issue.id,
      key: issue.key,
      summary: issue.fields.summary,
      status: issue.fields.status?.name,
      priority: issue.fields.priority?.name,
      assignee: issue.fields.assignee ? issue.fields.assignee.displayName : 'Unassigned',
      reporter: issue.fields.reporter?.displayName,
      created: issue.fields.created,
    }));
  } catch (error) {
    throw new Error(`Error fetching issues: ${error.message}`);
  }
}

export async function listIssues(client, projectKeyOrId = null) {
  try {
    const jql = projectKeyOrId ? `project = ${projectKeyOrId}` : '';
    return await searchIssues(client, jql)
  } catch (error) {
    throw new Error(`Error fetching issues ${projectKeyOrId ? `for project ${projectKeyOrId}` : ''}: ${error.message}`);
  }
}

export async function getIssue(client, issueIdOrKey) {
  try {
    const { data: issue } = await client.get(`/issue/${issueIdOrKey}`)

    return {
      id: issue.id,
      key: issue.key,
      summary: issue.fields.summary,
      status: issue.fields.status.name,
      description: issue.fields.description,
      priority: issue.fields.priority?.name,
      assignee: issue.fields.assignee ? issue.fields.assignee.displayName : 'Unassigned',
      reporter: issue.fields.reporter?.displayName,
      created: issue.fields.created,
    };
  } catch (error) {
    throw new Error(`Error fetching issue ${issueIdOrKey}: ${error.message}`);
  }
}

export async function createIssue(
  client,
  projectId,
  summary,
  description,
  issueTypeId,
  priorityId,
  assigneeId,
  reporterId
) {
  try {
    if (!projectId) {
      throw new Error('project_id argument is required');
    }
    if (!issueTypeId) {
      throw new Error('issue_type_id argument is required');
    }

    const body = {
      fields: {
        project: { id: projectId },
        issuetype: { id: issueTypeId },
        summary,
      },
    };

    if (priorityId) body.fields.priority = { id: priorityId };
    if (assigneeId) body.fields.assignee = { id: assigneeId };
    if (reporterId) body.fields.reporter = { id: reporterId };
    if (description) body.fields.description = JSON.parse(description);

    const { data: issue } = await client.post('/issue', body, {
        headers: {
            'Content-Type': 'application/json'
        }
    })

    return issue;
  } catch (error) {
    throw new Error(`Error creating issue: ${error.message}`);
  }
}

export async function editIssue(
  client,
  issueIdOrKey,
  newSummary,
  newDescription,
  newAssigneeId,
  newPriorityId,
  newName,
  newStatusName
) {
  try {
    // Construct the fields object based on the provided parameters
    const updatedFields = {};
    if (newSummary) updatedFields.summary = newSummary;
    if (newDescription) {
      updatedFields.description = JSON.parse(newDescription);
    }
    if (newAssigneeId) updatedFields.assignee = { id: newAssigneeId };
    if (newPriorityId) updatedFields.priority = { id: newPriorityId };
    if (newName) updatedFields.name = newName;

    // Update fields if any are specified
    if (Object.keys(updatedFields).length > 0) {
      await client.put(`/issue/${issueIdOrKey}`, { fields: updatedFields }, {
        headers: {
          'Content-Type': 'application/json',
        },
      });
    }

    // Handle status transition if specified
    if (newStatusName) {
      // Fetch available transitions
      const { data } = await client.get(`/issue/${issueIdOrKey}/transitions`);
      const transitions = data.transitions;

      // Find the transition ID matching the desired status name
      const transition = transitions.find(t => t.name.toLowerCase() === newStatusName.toLowerCase());
      if (!transition) {
        throw new Error(`Transition to status "${newStatusName}" not found for issue ${issueIdOrKey}.`);
      }

      // Execute the transition
      const transitionPayload = { transition: { id: transition.id } };
      await client.post(`/issue/${issueIdOrKey}/transitions`, transitionPayload, {
        headers: {
          'Content-Type': 'application/json',
        },
      });
    }

    return `Issue ${issueIdOrKey} updated successfully.`;
  } catch (error) {
    throw new Error(`Error editing issue ${issueIdOrKey}: ${error.message}`);
  }
}
