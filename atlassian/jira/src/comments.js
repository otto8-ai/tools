export async function addComment(client, issueIdOrKey, body) {
    try {
      if (!issueIdOrKey) {
        throw new Error('issue_id_or_key argument is required')
      }
      if (!body) {
        throw new Error('body argument is required')
      }

      // Parse the comment body JSON string into a JavaScript object
      const bodyADF = JSON.parse(body)

      // Validate that the parsed body is in ADF format
      if (!bodyADF || bodyADF.type !== 'doc' || !Array.isArray(bodyADF.content)) {
        throw new Error('Invalid comment body format. Expected Atlassian Document Format (ADF).')
      }

      const { data } = await client.post(`/issue/${issueIdOrKey}/comment`, { body: bodyADF }, {
        headers: {
          'Content-Type': 'application/json',
        },
      })

      return {
        id: data.id,
        author: data.author.displayName,
        created: data.created,
        updated: data.updated,
      }
    } catch (error) {
      throw new Error(`Error adding comment to issue ${issueIdOrKey}: ${error.message}`);
    }
  }

export async function listComments(client, issueIdOrKey) {
  try {
    if (!issueIdOrKey) {
      throw new Error('issue_id_or_key argument is required')
    }

    const { data } = await client.get(`/issue/${issueIdOrKey}/comment`)
    const comments = data.comments

    // Map the comments to include only the important fields
    return comments.map(comment => ({
      id: comment.id,
      author: comment.author?.displayName || 'Unknown',
      authorId: comment.author?.id,
      body: comment.body,
      created: comment.created,
      updated: comment.updated,
    }));
  } catch (error) {
    throw new Error(`Error fetching comments for issue ${issueIdOrKey}: ${error.message}`);
  }
}
