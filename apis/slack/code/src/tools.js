export async function listChannels(webClient) {
    const channels = await webClient.conversations.list({limit: 100, types: 'public_channel'})
    console.log('Public channels:')
    channels.channels.forEach(channel => {
        let printStr = `${channel.name} (ID: ${channel.id})`
        if (channel.is_archived === true) {
            printStr += ' (archived)'
        }
        console.log(printStr)
    })
    console.log('')

    const privateChannels = await webClient.conversations.list({limit: 100, types: 'private_channel'})
    console.log('Private channels:')
    privateChannels.channels.forEach(channel => {
        let printStr = `${channel.name} (ID: ${channel.id})`
        if (channel.is_archived === true) {
            printStr += ' (archived)'
        }
        console.log(printStr)
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
        console.log(`  message ID: ${message.ts}`)
        if (message.reply_count > 0) {
            console.log(`  thread ID ${threadID(message)} - ${message.reply_count} ${replyString(message.reply_count)}:`)
            const replies = await webClient.conversations.replies({channel: channelId, ts: message.ts, limit: 10})
            for (const reply of replies.messages) {
                if (reply.ts === message.ts) {
                    continue
                }

                const replyTime = new Date(parseFloat(reply.ts) * 1000)
                console.log(`  ${replyTime.toLocaleString()}: ${userMap[reply.user]}: ${reply.text}`)
                console.log(`    message ID: ${reply.ts}`)
            }
            if (replies.has_more) {
                console.log('  More replies exist')
            }
        }
    }
}

export async function getThreadHistory(webClient, channelId, threadId, limit) {
    const replies = await webClient.conversations.replies({channel: channelId, ts: threadId, limit: limit})
    if (!replies.ok) {
        console.error(`Failed to retrieve thread history: ${replies.error}`)
        process.exit(1)
    } else if (replies.messages.length === 0) {
        console.log('No messages found')
        return
    }

    const userMap = await getUserMap(webClient)

    for (const reply of replies.messages) {
        const time = new Date(parseFloat(reply.ts) * 1000)
        console.log(`${time.toLocaleString()}: ${userMap[reply.user]}: ${reply.text}`)
        console.log(`  message ID: ${reply.ts}`)
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
        console.log(`  message ID: ${message.ts}`)
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

export async function sendMessageInThread(webClient, channelId, threadTs, text) {
    const result = await webClient.chat.postMessage({
        channel: channelId,
        text: text,
        thread_ts: threadTs,
    })

    if (!result.ok) {
        console.error(`Failed to send message: ${result.error}`)
        process.exit(1)
    }
    console.log('Thread message sent successfully')
}

export async function listUsers(webClient) {
    const users = await webClient.users.list()
    users.members.forEach(user => {
        console.log(user.name)
        console.log(`  ID: ${user.id}`)
        console.log(`  Full name: ${user.profile.real_name}`)
        console.log(`  Account deleted: ${user.deleted}`)
    })
}

export async function searchUsers(webClient, query) {
    const users = await webClient.users.list()
    users.members.forEach(user => {
        if (user.name.includes(query) || user.profile.real_name.includes(query)) {
            console.log(user.name)
            console.log(`  ID: ${user.id}`)
            console.log(`  Full name: ${user.profile.real_name}`)
            console.log(`  Account deleted: ${user.deleted}`)
        }
    })
}

export async function sendDM(webClient, userId, text) {
    const res = await webClient.conversations.open({
        users: userId,
    })

    await webClient.chat.postMessage({
        channel: res.channel.id,
        text,
    })

    console.log('Message sent successfully')
}

export async function getMessageLink(webClient, channelId, messageId) {
    const result = await webClient.chat.getPermalink({
        channel: channelId,
        message_ts: messageId,
    })

    if (!result.ok) {
        console.error(`Failed to get message link: ${result.error}`)
        process.exit(1)
    }

    console.log(result.permalink)
}

// Helper functions below

function replyString(count) {
    return count === 1 ? 'reply' : 'replies'
}

function threadID(message) {
    return message.ts
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
