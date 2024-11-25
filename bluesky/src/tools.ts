import { AtpAgent } from '@atproto/api'
import { createPost, deletePost, searchPosts } from './posts.ts'
import { searchUsers } from './users.ts'

if (process.argv.length !== 3) {
  console.error('Usage: node tool.ts <command>')
  process.exit(1)
}

const BLUESKY_HANDLE = process.env.BLUESKY_HANDLE
const BLUESKY_APP_PASSWORD = process.env.BLUESKY_APP_PASSWORD

const command = process.argv[2]

try {
  if (!BLUESKY_HANDLE) {
    throw new Error('bluesky username not set')
  }

  if (!BLUESKY_APP_PASSWORD) {
    throw new Error('bluesky app password not set')
  }

  const agent = new AtpAgent({
    service: 'https://bsky.social'
  })

  await agent.login({
    identifier: BLUESKY_HANDLE,
    password: BLUESKY_APP_PASSWORD
  })

  switch (command) {
      case 'createPost':
          await createPost(
              agent,
              process.env.TEXT,
              process.env.TAGS,
          )
          break
      case 'deletePost':
          await deletePost(
              agent,
              process.env.POST_URI,
          )
          break
      case 'searchPosts':
          await searchPosts(
              agent,
              process.env.QUERY,
              process.env.SINCE,
              process.env.UNTIL,
              process.env.LIMIT,
              process.env.TAGS,
          )
          break
      case 'searchUsers':
          await searchUsers(
              agent,
              process.env.QUERY,
              process.env.LIMIT,
          )
          break
      default:
          console.log(`Unknown command: ${command}`)
          process.exit(1)
  }
} catch (error: unknown) {
  // Print the error to stdout so that it can be captured by the GPTScript
  console.log(String(error))
  process.exit(1)
}
