import { AtpAgent, AppBskyFeedSearchPosts } from '@atproto/api'

export async function searchPosts (
    agent: AtpAgent,
    query?: string,
    since?: string,
    until?: string,
    limit?: string,
    tags?: string,
): Promise<void> {
    let queryParams: AppBskyFeedSearchPosts.QueryParams = {
        q: query ?? '',
        sort: 'latest',
        limit: 25
    }

    if (!query) {
        throw new Error('Query is required')
    }

    if (!!limit) {
        try {
            queryParams.limit = parseInt(limit, 10)
        } catch (error: unknown) {
            throw new Error(`Invalid limit format: ${String(error)}`)
        }
    }

    if (!!until) {
        try {
            queryParams.until = new Date(until).toISOString()
        } catch (error: unknown) {
            throw new Error(`Invalid until date format: ${String(error)}`)
        }
    }

    if (!!since) {
        try {
            queryParams.since = new Date(since).toISOString()
        } catch (error: unknown) {
            throw new Error(`Invalid since date format: ${String(error)}`)
        }
    }

    if (!!tags) {
        queryParams.tag = tags
            .split(',')
            .map(tag => tag.trim().replace(/^#/, ''))
    }

    const response = await agent.app.bsky.feed.searchPosts(queryParams)

    console.log(JSON.stringify(response.data.posts))
}

export async function createPost(agent: AtpAgent, text?: string, tags?: string): Promise<void> {
    if (!text) {
        throw new Error('Text is required')
    }

    await agent.post({
        text,
        tags: tags?.split(',').map(tag => tag.trim().replace(/^#/, '')) ?? [],
    })

    console.log('Post created')
}

export async function deletePost(agent: AtpAgent, postUri?: string): Promise<void> {
    if (!postUri) {
        throw new Error('Post URI is required')
    }

    await agent.deletePost(postUri)

    console.log('Post deleted')
}
