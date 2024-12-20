export async function listPriorities(client, schemeId) {
  try {
    let schemeIds = []
    if (schemeId) {
      schemeIds = [schemeId]
    } else {
      const schemeIdList = await listPrioritySchemes(client);
      schemeIds = schemeIdList.values?.map(scheme => scheme.id) || []
    }

    if (schemeIds.length < 1) {
      throw new Error('No priority schemes found.');
    }

    let priorities = []
    for (const schemeId of schemeIds) {
      const { data: schemePriorities } = await client.get(`/priorityscheme/${schemeId}/priorities`)
      priorities = [...priorities, ...schemePriorities.values]
    }

    return priorities
  } catch (error) {
    throw new Error(`Error fetching priorities for scheme ${schemeId}: ${error.message}`);
  }
}

async function listPrioritySchemes(client) {
  try {
    const { data } = await client.get('/priorityscheme')
    const prioritySchemes = data.values

    return prioritySchemes.map(scheme => ({
      id: scheme.id,
      name: scheme.name,
      priorities: scheme.priorities.values.map(priority => ({
        id: priority.id,
        name: priority.name,
        description: priority.description,
        isDefault: priority.isDefault,
      })),
    }));
  } catch (error) {
    throw new Error(`Error fetching priority schemes: ${error.message}`);
  }
}