export async function listProjects(client) {
  try {
    const { data: projects } = await client.get('/project')

    return projects.map(project => ({
      id: project.id,
      key: project.key,
      name: project.name,
    }))
  } catch (error) {
    throw new Error(`Error fetching projects: ${error.message}`);
  }
}

export async function getProject(client, projectIdOrKey) {
  try {
    if (!projectIdOrKey) {
      throw new Error('project_id_or_key argument is required');
    }

    // Fetch project details and issue types with statuses
    const { data: project } = await client.get(`/project/${projectIdOrKey}`);
    const { data: issueTypesWithStatuses } = await client.get(`/project/${projectIdOrKey}/statuses`);

    // Create a Map for fast lookup by id
    const statusesMap = new Map(issueTypesWithStatuses.map(status => [status.id, status]));

    // Merge issue types with statuses using the Map
    const issueTypesWithMergedStatuses = project?.issueTypes?.map(issueType => ({
      ...issueType,
      statuses: statusesMap.get(issueType.id)?.statuses || [], // Merge statuses if available
    }));

    return {
      id: project.id,
      key: project.key,
      name: project.name,
      projectUrl: project?.self,
      description: project?.description,
      projectType: project?.projectTypeKey,
      deleted: project?.deleted,
      archived: project?.archived,
      assigneeType: project?.assigneeType,
      style: project?.style,
      lead: {
        accountId: project.lead?.accountId,
        active: project.lead?.active,
      },
      avatarUrl: project?.avatarUrls['48x48'],
      issueTypes: issueTypesWithMergedStatuses, // Use the merged issue types
      issueTypeHierarchy: project?.issueTypeHierarchy,
      roles: project?.roles,
      projectTypeKey: project?.projectTypeKey,
    };
  } catch (error) {
    throw new Error(`Error fetching project ${projectIdOrKey}: ${error.message}`);
  }
}
