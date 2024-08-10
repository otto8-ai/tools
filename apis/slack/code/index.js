import {WebClient} from '@slack/web-api'
import {getChannelHistory, listChannels, search, sendMessage} from "./src/tools.js";

if (process.argv.length !== 3) {
    console.error('Usage: node index.js <command>')
    process.exit(1)
}

const command = process.argv[2]
const token = process.env.SLACK_TOKEN

const webClient = new WebClient(token)

switch (command) {
    case "listChannels":
        await listChannels(webClient)
        break
    case "getChannelHistory":
        await getChannelHistory(webClient, process.env.CHANNELID, process.env.LIMIT)
        break
    case "searchMessages":
        await search(webClient, process.env.QUERY)
        break
    case "sendMessage":
        await sendMessage(webClient, process.env.CHANNELID, process.env.TEXT)
        break
    default:
        console.error(`Unknown command: ${command}`)
        process.exit(1)
}
