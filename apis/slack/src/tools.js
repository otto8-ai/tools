export async function listChannels(webClient) {
    const channels = await webClient.conversations.list({limit: 100, types: 'public_channel'})
    console.log('Public channels:')
    channels.channels.forEach(channel => {
        printChannel(channel)
    })
    console.log('')

    const privateChannels = await webClient.conversations.list({limit: 100, types: 'private_channel'})
    console.log('Private channels:')
    privateChannels.channels.forEach(channel => {
        printChannel(channel)
    })
}

export async function searchChannels(webClient, query) {
    const channels = await webClient.conversations.list({limit: 100, types: 'public_channel'})
    channels.channels.forEach(channel => {
        if (channel.name.includes(query)) {
            printChannel(channel)
        }
    })

    const privateChannels = await webClient.conversations.list({limit: 100, types: 'private_channel'})
    privateChannels.channels.forEach(channel => {
        if (channel.name.includes(query)) {
            printChannel(channel)
        }
    })
}

export async function getChannelHistory(webClient, channelId, limit) {
    const history = await webClient.conversations.history({channel: channelId, limit: limit})
    if (!history.ok) {
        console.log(`Failed to retrieve chat history: ${history.error}`)
        process.exit(1)
    } else if (history.messages.length === 0) {
        console.log('No messages found')
        return
    }

    await printHistory(webClient, channelId, history)
}

export async function getChannelHistoryByTime(webClient, channelId, limit, start, end) {
    const oldest = new Date(start).getTime() / 1000
    const latest = new Date(end).getTime() / 1000
    const history = await webClient.conversations.history({channel: channelId, limit: limit, oldest: oldest.toString(), latest: latest.toString()})
    if (!history.ok) {
        console.log(`Failed to retrieve chat history: ${history.error}`)
        process.exit(1)
    } else if (history.messages.length === 0) {
        console.log('No messages found')
        return
    }

    await printHistory(webClient, channelId, history)
}

export async function getThreadHistory(webClient, channelId, threadId, limit) {
    const replies = await webClient.conversations.replies({channel: channelId, ts: threadId, limit: limit})
    if (!replies.ok) {
        console.log(`Failed to retrieve thread history: ${replies.error}`)
        process.exit(1)
    } else if (replies.messages.length === 0) {
        console.log('No messages found')
        return
    }

    for (const reply of replies.messages) {
        await printMessage(webClient, reply)
    }
}


export async function search(webClient, query) {
    const result = await webClient.search.all({
        query: query,
    })

    if (!result.ok) {
        console.log(`Failed to search messages: ${result.error}`)
        process.exit(1)
    }

    if (result.messages.matches.length === 0) {
        console.log('No messages found')
        return
    }

    for (const message of result.messages.matches) {
        await printMessage(webClient, message)
    }
}

export async function sendMessage(webClient, channelId, text) {
    const result = await webClient.chat.postMessage({
        channel: channelId,
        text: text,
    })

    if (!result.ok) {
        console.log(`Failed to send message: ${result.error}`)
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
        console.log(`Failed to send message: ${result.error}`)
        process.exit(1)
    }
    console.log('Thread message sent successfully')
}

export async function listUsers(webClient) {
    const users = await webClient.users.list()
    users.members.forEach(user => {
        printUser(user)
    })
}

export async function searchUsers(webClient, query) {
    const users = await webClient.users.list()
    users.members.forEach(user => {
        if (user.name.includes(query) || user.profile.real_name.includes(query)) {
            printUser(user)
        }
    })
}

export async function sendDM(webClient, userIds, text) {
    const res = await webClient.conversations.open({
        users: userIds,
    })

    await webClient.chat.postMessage({
        channel: res.channel.id,
        text,
    })

    console.log('Message sent successfully')
}

export async function sendDMInThread(webClient, userIds, threadId, text) {
    const res = await webClient.conversations.open({
        users: userIds,
    })

    await webClient.chat.postMessage({
        channel: res.channel.id,
        text,
        thread_ts: threadId,
    })

    console.log('Thread message sent successfully')
}

export async function getMessageLink(webClient, channelId, messageId) {
    const result = await webClient.chat.getPermalink({
        channel: channelId,
        message_ts: messageId,
    })

    if (!result.ok) {
        console.log(`Failed to get message link: ${result.error}`)
        process.exit(1)
    }

    console.log(result.permalink)
}

export async function getDMHistory(webClient, userIds, limit) {
    const res = await webClient.conversations.open({
        users: userIds,
    })

    const history = await webClient.conversations.history({
        channel: res.channel.id,
        limit: limit,
    })

    if (!history.ok) {
        console.log(`Failed to retrieve chat history: ${history.error}`)
        process.exit(1)
    }

    if (history.messages.length === 0) {
        console.log('No messages found')
        return
    }

    for (const message of history.messages) {
        await printMessage(webClient, message)
    }
}

export async function getDMThreadHistory(webClient, userIds, threadId, limit) {
    const res = await webClient.conversations.open({
        users: userIds,
    })

    const replies = await webClient.conversations.replies({
        channel: res.channel.id,
        ts: threadId,
        limit: limit,
    })

    if (!replies.ok) {
        console.log(`Failed to retrieve thread history: ${replies.error}`)
        process.exit(1)
    }

    if (replies.messages.length === 0) {
        console.log('No messages found')
        return
    }

    for (const reply of replies.messages) {
        await printMessage(webClient, reply)
    }
}

// Helper functions below

function replyString(count) {
    return count === 1 ? 'reply' : 'replies'
}

function threadID(message) {
    return message.ts
}

async function getUserName(webClient, user) {
    const res = await webClient.users.info({user: user})
    if (!res.ok) {
        // If the request didn't work for some reason, just return the user ID again.
        return user
    }
    return res.user.name
}

// Printer functions below

function printUser(user) {
    console.log(user.name)
    console.log(`  ID: ${user.id}`)
    console.log(`  Full name: ${user.profile.real_name}`)
    if (user.deleted === true) {
        console.log(`  Account deleted: true`)
    }
}

async function printMessage(webClient, message) {
    const time = new Date(parseFloat(message.ts) * 1000)
    let userName = message.user
    try {
        userName = await getUserName(webClient, message.user)
    } catch (e) {}

    console.log(`${time.toLocaleString()}: ${userName}: ${message.text}`)
    console.log(`  message ID: ${message.ts}`)
    if (message.blocks && message.blocks.length > 0) {
        console.log(`  message blocks: ${JSON.stringify(message.blocks)}`)
    }
    if (message.attachments && message.attachments.length > 0) {
        console.log(`  message attachments: ${JSON.stringify(message.attachments)}`)
    }
}

function printChannel(channel) {
    let printStr = `${channel.name} (ID: ${channel.id})`
    if (channel.is_archived === true) {
        printStr += ' (archived)'
    }
    console.log(printStr)
}

async function printHistory(webClient, channelId, history) {
    for (const message of history.messages) {
        await printMessage(webClient, message)
        if (message.reply_count > 0) {
            console.log(`  thread ID ${threadID(message)} - ${message.reply_count} ${replyString(message.reply_count)}:`)
            const replies = await webClient.conversations.replies({channel: channelId, ts: message.ts, limit: 3})
            for (const reply of replies.messages) {
                if (reply.ts === message.ts) {
                    continue
                }

                await printMessage(webClient, reply)
            }
            if (replies.has_more) {
                console.log('  More replies exist')
            }
        }
    }
}
