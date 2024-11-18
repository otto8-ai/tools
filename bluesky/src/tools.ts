import { AtpAgent } from '@atproto/api'
import { searchPosts } from './posts.ts'
import { searchUsers } from './users.ts'

if (process.argv.length !== 3) {
  console.error('Usage: node tool.ts <command>')
  process.exit(1)
}

const command = process.argv[2]

try {
  const agent = new AtpAgent({
    service: 'https://api.bsky.app'
  })

  switch (command) {
      case 'searchPosts':
          await searchPosts(
              agent,
              process.env.QUERY,
              process.env.SINCE,
              process.env.UNTIL,
              process.env.LIMIT,
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
