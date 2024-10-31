import { WebClient } from "@slack/web-api"
import {
  getChannelHistory,
  getChannelHistoryByTime,
  getDMHistory,
  getDMThreadHistory,
  getMessageLink,
  getThreadHistory,
  listChannels,
  listUsers,
  search,
  searchChannels,
  searchUsers,
  sendDM,
  sendDMInThread,
  sendMessage,
  sendMessageInThread,
  userContext,
} from "./src/tools.js"

if (process.argv.length !== 3) {
  console.error("Usage: node index.js <command>")
  process.exit(1)
}

const command = process.argv[2]
const token = process.env.SLACK_TOKEN

const webClient = new WebClient(token)

switch (command) {
  case "listChannels":
    await listChannels(webClient)
    break
  case "searchChannels":
    await searchChannels(webClient, process.env.QUERY)
    break
  case "getChannelHistory":
    await getChannelHistory(webClient, process.env.CHANNELID, process.env.LIMIT)
    break
  case "getChannelHistoryByTime":
    await getChannelHistoryByTime(webClient, process.env.CHANNELID, process.env.LIMIT, process.env.START, process.env.END)
    break
  case "getThreadHistory":
    await getThreadHistory(webClient, process.env.CHANNELID, process.env.THREADID, process.env.LIMIT)
    break
  case "searchMessages":
    await search(webClient, process.env.QUERY)
    break
  case "sendMessage":
    await sendMessage(webClient, process.env.CHANNELID, process.env.TEXT)
    break
  case "sendMessageInThread":
    await sendMessageInThread(webClient, process.env.CHANNELID, process.env.THREADID, process.env.TEXT)
    break
  case "listUsers":
    await listUsers(webClient)
    break
  case "searchUsers":
    await searchUsers(webClient, process.env.QUERY)
    break
  case "sendDM":
    await sendDM(webClient, process.env.USERIDS, process.env.TEXT)
    break
  case "sendDMInThread":
    await sendDMInThread(webClient, process.env.USERIDS, process.env.THREADID, process.env.TEXT)
    break
  case "getMessageLink":
    await getMessageLink(webClient, process.env.CHANNELID, process.env.MESSAGEID)
    break
  case "getDMHistory":
    await getDMHistory(webClient, process.env.USERIDS, process.env.LIMIT)
    break
  case "getDMThreadHistory":
    await getDMThreadHistory(webClient, process.env.USERIDS, process.env.THREADID, process.env.LIMIT)
    break
  case "userContext":
    await userContext(webClient)
    break
  default:
    console.error(`Unknown command: ${command}`)
    process.exit(1)
}
