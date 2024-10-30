import {GPTScript} from "@gptscript-ai/gptscript";

export async function listChannels(webClient) {
    const publicChannels = await webClient.conversations.list({limit: 100, types: 'public_channel'})
    const privateChannels = await webClient.conversations.list({limit: 100, types: 'private_channel'})

    if (publicChannels.channels.length + privateChannels.channels.length > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, 'slack_channels', 'list of slack channels')

            for (const channel of [...publicChannels.channels, ...privateChannels.channels]) {
                await gptscriptClient.addDatasetElement(process.env.GPTSCRIPT_WORKSPACE_ID, dataset.id, channel.name, channel.purpose.value || '', channelToString(channel))
            }

            console.log(`Created dataset with ID ${dataset.id} with ${publicChannels.channels.length + privateChannels.channels.length} channels`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    console.log('Public channels:')
    publicChannels.channels.forEach(channel => {
        console.log(channelToString(channel))
    })
    console.log('')


    console.log('Private channels:')
    privateChannels.channels.forEach(channel => {
        console.log(channelToString(channel))
    })
}

export async function searchChannels(webClient, query) {
    const publicResult = await webClient.conversations.list({limit: 100, types: 'public_channel'})
    const privateResult = await webClient.conversations.list({limit: 100, types: 'private_channel'})

    const publicChannels = publicResult.channels.filter(channel => channel.name.includes(query))
    const privateChannels = privateResult.channels.filter(channel => channel.name.includes(query))

    if (publicChannels.length + privateChannels.length === 0) {
        console.log('No channels found')
        return
    } else if (publicChannels.length + privateChannels.length > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, `${query}_slack_channels`, `list of slack channels matching search query "${query}"`)

            for (const channel of [...publicChannels, ...privateChannels]) {
                await gptscriptClient.addDatasetElement(
                    process.env.GPTSCRIPT_WORKSPACE_ID,
                    dataset.id,
                    channel.name,
                    channel.purpose.value || '',
                    channelToString(channel))
            }

            console.log(`Created dataset with ID ${dataset.id} with ${publicChannels.length + privateChannels.length} channels`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    publicChannels.forEach(channel => {
        if (channel.name.includes(query)) {
            console.log(channelToString(channel))
        }
    })
    console.log("")
    privateChannels.forEach(channel => {
        if (channel.name.includes(query)) {
            console.log(channelToString(channel))
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

    if (replies.messages.length > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, `slack_thread_${threadId}`, `thread history for thread "${threadId}"`)

            for (const reply of replies.messages) {
                await gptscriptClient.addDatasetElement(process.env.GPTSCRIPT_WORKSPACE_ID, dataset.id, reply.ts, "", await messageToString(webClient, reply))
            }

            console.log(`Created dataset with ID ${dataset.id} with ${replies.messages.length} thread replies`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    for (const reply of replies.messages) {
        console.log(await messageToString(webClient, reply))
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
    } else if (result.messages.matches.length > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, `slack_search_${query}`, `search results for query "${query}"`)

            for (const message of result.messages.matches) {
                await gptscriptClient.addDatasetElement(
                    process.env.GPTSCRIPT_WORKSPACE_ID,
                    dataset.id,
                    `${message.iid}_${message.ts}`,
                    "",
                    await messageToString(webClient, message)
                )
            }

            console.log(`Created dataset with ID ${dataset.id} with ${result.messages.matches.length} search results`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    for (const message of result.messages.matches) {
        console.log(await messageToString(webClient, message))
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

    if (users.members.length > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, 'slack_users', 'list of slack users')

            for (const user of users.members) {
                await gptscriptClient.addDatasetElement(process.env.GPTSCRIPT_WORKSPACE_ID, dataset.id, user.name, user.profile.real_name, userToString(user))
            }

            console.log(`Created dataset with ID ${dataset.id} with ${users.members.length} users`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    users.members.forEach(user => {
        console.log(userToString(user))
    })
}

export async function searchUsers(webClient, query) {
    const users = await webClient.users.list()
    const matchingUsers = users.members.filter(user => user.name.includes(query) || user.profile.real_name.includes(query))

    if (matchingUsers.length > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, `${query}_slack_users`, `list of slack users matching search query "${query}"`)

            for (const user of matchingUsers) {
                await gptscriptClient.addDatasetElement(process.env.GPTSCRIPT_WORKSPACE_ID, dataset.id, user.name, user.profile.real_name, userToString(user))
            }

            console.log(`Created dataset with ID ${dataset.id} with ${matchingUsers.length} users`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    matchingUsers.forEach(user => {
        if (user.name.includes(query) || user.profile.real_name.includes(query)) {
            console.log(userToString(user))
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
    } else if (history.messages.length > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, `slack_dm_history_${userIds}`, `chat history for DM with users "${userIds}"`)

            for (const message of history.messages) {
                await gptscriptClient.addDatasetElement(process.env.GPTSCRIPT_WORKSPACE_ID, dataset.id, message.ts, "", await messageToString(webClient, message))
            }

            console.log(`Created dataset with ID ${dataset.id} with ${history.messages.length} messages`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    for (const message of history.messages) {
        console.log(await messageToString(webClient, message))
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
    } else if (replies.messages.length > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, `slack_dm_thread_${threadId}`, `thread history for DM with users "${userIds}"`)

            for (const reply of replies.messages) {
                await gptscriptClient.addDatasetElement(process.env.GPTSCRIPT_WORKSPACE_ID, dataset.id, reply.ts, "", await messageToString(webClient, reply))
            }

            console.log(`Created dataset with ID ${dataset.id} with ${replies.messages.length} thread replies`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    for (const reply of replies.messages) {
        console.log(await messageToString(webClient, reply))
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

function userToString(user) {
    let str = `${user.name}`
    str += `  ID: ${user.id}`
    str += `  Full name: ${user.profile.real_name}`
    if (user.deleted === true) {
        str += '  Account deleted: true'
    }
    return str
}

async function messageToString(webClient, message) {
    const time = new Date(parseFloat(message.ts) * 1000)
    let userName = message.user
    try {
        userName = await getUserName(webClient, message.user)
    } catch (e) {}

    // Find and replace any user mentions in the message text with the user's name
    const userMentions = message.text.match(/<@U[A-Z0-9]+>/g) ?? []
    for (const mention of userMentions) {
        const userId = mention.substring(2, mention.length - 1)
        try {
            const userName = await getUserName(webClient, userId)
            message.text = message.text.replace(mention, `@${userName}`)
        } catch (e) {}
    }

    let str = `${time.toLocaleString()}: ${userName}: ${message.text}\n`
    str += `  message ID: ${message.ts}\n`
    if (message.blocks && message.blocks.length > 0) {
        str += `  message blocks: ${JSON.stringify(message.blocks)}\n`
    }
    if (message.attachments && message.attachments.length > 0) {
        str += `  message attachments: ${JSON.stringify(message.attachments)}\n`
    }
    return str
}

function channelToString(channel) {
    let str = `${channel.name} (ID: ${channel.id})`
    if (channel.is_archived === true) {
        str += ' (archived)'
    }
    return str
}

async function printHistory(webClient, channelId, history) {
    const data = new Map()

    for (const message of history.messages) {
        let messageStr = await messageToString(webClient, message)
        if (message.reply_count > 0) {
            messageStr += `\n  thread ID ${threadID(message)} - ${message.reply_count} ${replyString(message.reply_count)}:`
            const replies = await webClient.conversations.replies({channel: channelId, ts: message.ts, limit: 3})
            for (const reply of replies.messages) {
                if (reply.ts === message.ts) {
                    continue
                }

                messageStr += "\n" + await messageToString(webClient, reply)
            }
            if (replies.has_more) {
                messageStr += '\n  More replies exist'
            }
        }

        data.set(message.ts, messageStr)
    }

    if (data.size > 10) {
        try {
            const gptscriptClient = new GPTScript()
            const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_ID, `slack_history_${channelId}`, `chat history for channel "${channelId}"`)

            for (const [key, value] of data.entries()) {
                await gptscriptClient.addDatasetElement(process.env.GPTSCRIPT_WORKSPACE_ID, dataset.id, key, "", value)
            }

            console.log(`Created dataset with ID ${dataset.id} with ${data.size} messages`)
            return
        } catch (e) {} // Ignore errors if we got any. We'll just print the results below.
    }

    for (const [key, value] of data.entries()) {
        console.log(value)
    }
}
