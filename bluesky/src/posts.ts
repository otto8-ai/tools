import { AtpAgent, AppBskyFeedSearchPosts } from '@atproto/api'

export async function searchPosts (
    agent: AtpAgent,
    query?: string,
    since?: string,
    until?: string,
    limit?: string,
): Promise<void> {
    let queryParams: AppBskyFeedSearchPosts.QueryParams = {
        q: query ?? '',
        sort: 'latest',
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

    const response = await agent.app.bsky.feed.searchPosts(queryParams)

    console.log(JSON.stringify(response.data.posts))
}