export async function listChannels(webClient) {
    const channels = await webClient.conversations.list({limit: 100})
    channels.channels.forEach(channel => {
        console.log(`${channel.name} (ID: ${channel.id})`)
    })
}

export async function getChannelHistory(webClient, channelId, limit) {
    const history = await webClient.conversations.history({channel: channelId, limit: limit})
    if (!history.ok) {
        console.error(`Failed to retrieve chat history: ${history.error}`)
        process.exit(1)
    } else if (history.messages.length === 0) {
        console.log('No messages found')
        return
    }

    const userMap = await getUserMap(webClient)

    for (const message of history.messages) {
        const time = new Date(parseFloat(message.ts) * 1000)
        console.log(`${time.toLocaleString()}: ${userMap[message.user]}: ${message.text}`)
        if (message.reply_count > 0) {
            console.log(`  ${message.reply_count} ${replyString(message.reply_count)}:`)
            const replies = await webClient.conversations.replies({channel: channelId, ts: message.ts})
            for (const reply of replies.messages) {
                if (reply.ts === message.ts) {
                    continue
                }

                const replyTime = new Date(parseFloat(reply.ts) * 1000)
                console.log(`  ${replyTime.toLocaleString()}: ${userMap[reply.user]}: ${reply.text}`)
            }
        }
    }
}


export async function search(webClient, query) {
    const result = await webClient.search.all({
        query: query,
    })

    if (!result.ok) {
        console.error(`Failed to search messages: ${result.error}`)
        process.exit(1)
    }

    if (result.messages.matches.length === 0) {
        console.log('No messages found')
        return
    }

    const userMap = await getUserMap(webClient)

    for (const message of result.messages.matches) {
        const time = new Date(parseFloat(message.ts) * 1000)
        console.log(`${time.toLocaleString()}: ${userMap[message.user]} in #${message.channel.name}: ${message.text}`)
    }
}

export async function sendMessage(webClient, channelId, text) {
    const result = await webClient.chat.postMessage({
        channel: channelId,
        text: text,
    })

    if (!result.ok) {
        console.error(`Failed to send message: ${result.error}`)
        process.exit(1)
    }
    console.log('Message sent successfully')
}

// Helper functions below

function replyString(count) {
    return count === 1 ? 'reply' : 'replies'
}

async function getUserMap(webClient) {
    // Get the list of users. We will need this in order to look up usernames.
    const users = await webClient.users.list()
    if (!users.ok) {
        console.error(`Failed to retrieve user list: ${users.error}`)
        process.exit(1)
    }

    // Create a map of user IDs to usernames.
    const userMap = {}
    users.members.forEach(user => {
        userMap[user.id] = user.name
    })

    return userMap
}
