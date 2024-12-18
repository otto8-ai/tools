const REQUIRED_SCOPES = ['write:jira-work', 'read:jira-work', 'read:jira-user'];
const hasRequiredScopes = (scopes) =>
  REQUIRED_SCOPES.every(item => scopes.includes(item));

export async function listJiraSites(client) {
  try {
    const { data } = await client.get('https://api.atlassian.com/oauth/token/accessible-resources')

    let jiraSites = [];
    for (const site of data) {
      if (hasRequiredScopes(site.scopes)) {
        const { scopes, ...rest } = site;
        jiraSites.push(rest);
      }
    }

    return jiraSites;
  } catch (error) {
    throw new Error(`Error fetching Jira sites: ${error.message}`);
  }
}
