import { GPTScript } from "@gptscript-ai/gptscript"
import { Mutex } from "async-mutex"

export async function userContext(webClient) {
    const result = await webClient.auth.test({})
    const userResult = await webClient.users.info({user: result.user_id})
    console.log(`Logged in as ${userResult.user.name}`)
    console.log(`Real Name: ${userResult.user.profile.real_name}`)
    console.log(`Display Name: ${userResult.user.profile.display_name}`)
    console.log(`User ID: ${result.user_id}`)
    console.log(`Default time zone: ${userResult.user.tz_label} (UTC offset ${userResult.user.tz_offset / 3600} hours)`)
}

export async function listChannels(webClient) {
    const publicChannels = await webClient.conversations.list({limit: 100, types: 'public_channel'})
    const privateChannels = await webClient.conversations.list({limit: 100, types: 'private_channel'})

    try {
        const gptscriptClient = new GPTScript()
        const elements = [...publicChannels.channels, ...privateChannels.channels].map(channel => {
            return {
                name: `${channel.name}`,
                description: `${channel.purpose.value || ''}`,
                contents: channelToString(channel)
            }
        })

        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: "slack_channels"
        })
        console.log(`Created dataset with ID ${datasetID} with ${publicChannels.channels.length + privateChannels.channels.length} channels`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
    }
}

export async function searchChannels(webClient, query) {
    const publicResult = await webClient.conversations.list({limit: 100, types: 'public_channel'})
    const privateResult = await webClient.conversations.list({limit: 100, types: 'private_channel'})

    const publicChannels = publicResult.channels.filter(channel => channel.name.includes(query))
    const privateChannels = privateResult.channels.filter(channel => channel.name.includes(query))

    if (publicChannels.length + privateChannels.length === 0) {
        console.log('No channels found')
        return
    }

    try {
        const gptscriptClient = new GPTScript()
        const elements = [...publicChannels, ...privateChannels].map(channel => {
            return {
                name: `${channel.name}`,
                description: `${channel.name} (ID: ${channel.id})`,
                contents: channelToString(channel)
            }
        })
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `${query}_slack_channels`,
            description: `list of slack channels matching search query "${query}"`
        })
        console.log(`Created dataset with ID ${datasetID} with ${publicChannels.length + privateChannels.length} channels`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
    }
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

    try {
        const gptscriptClient = new GPTScript()
        const elements = await Promise.all(replies.messages.map(async (reply) => {
            return {
                name: `${reply.ts}`,
                description: '',
                contents: await messageToString(webClient, reply)
            }
        }))
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `slack_thread_${threadId}`,
            description: `thread history for thread "${threadId}"`
        })
        console.log(`Created dataset with ID ${datasetID} with ${replies.messages.length} thread replies`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
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

    try {
        const gptscriptClient = new GPTScript()
        const elements = await Promise.all(result.messages.matches.map(async (message) => {
            return {
                name: `${message.iid}_${message.ts}`,
                description: '',
                contents: await messageToString(webClient, message)
            }
        }))
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `slack_search_${query}`,
            description: `search results for query "${query}"`
        })
        console.log(`Created dataset with ID ${datasetID} with ${result.messages.matches.length} search results`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
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

    try {
        const gptscriptClient = new GPTScript()
        const elements = users.members.map(user => {
            return {
                name: `${user.name}`,
                description: `${user.profile.real_name}`,
                contents: userToString(user)
            }
        })
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: "slack_users"
        })
        console.log(`Created dataset with ID ${datasetID} with ${users.members.length} users`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
    }
}

export async function searchUsers(webClient, query) {
    const users = await webClient.users.list()
    const matchingUsers = users.members.filter(user => {
        return user.name.includes(query) ||
            user.profile.real_name.includes(query) ||
            user.profile.display_name.includes(query)
    })
    if (matchingUsers.length === 0) {
        console.log('No users found')
        return
    }

    try {
        const gptscriptClient = new GPTScript()
        const elements = matchingUsers.map(user => {
            return {
                name: `${user.name}`,
                description: `${user.profile.real_name}`,
                contents: userToString(user)
            }
        })
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `${query}_slack_users`,
            description: `list of slack users matching search query "${query}"`
        })
        console.log(`Created dataset with ID ${datasetID} with ${matchingUsers.length} users`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
    }
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

    try {
        const gptscriptClient = new GPTScript()
        const elements = await Promise.all(history.messages.map(async (message) => {
            return {
                name: `${message.ts}`,
                description: '',
                contents: await messageToString(webClient, message)
            }
        }))
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `slack_dm_history_${userIds}`,
            description: `chat history for DM with users "${userIds}"`
        })
        console.log(`Created dataset with ID ${datasetID} with ${history.messages.length} messages`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
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

    try {
        const gptscriptClient = new GPTScript()
        const elements = await Promise.all(replies.messages.map(async (reply) => {
            return {
                name: `${reply.ts}`,
                description: '',
                contents: await messageToString(webClient, reply)
            }
        }))
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `slack_dm_thread_${threadId}`,
            description: `thread history for DM with users "${userIds}"`
        })
        console.log(`Created dataset with ID ${datasetID} with ${replies.messages.length} thread replies`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
    }
}

// Helper functions below

function replyString(count) {
    return count === 1 ? 'reply' : 'replies'
}

function threadID(message) {
    return message.ts
}

const userNameCache = new Map()
const userNameLock = new Mutex()

async function getUserName(webClient, user) {
    // Check if the username is already cached
    if (userNameCache.has(user)) {
        return userNameCache.get(user);
    }

    return await userNameLock.runExclusive(async () => {
        // Double-check the cache inside the lock
        if (userNameCache.has(user)) {
            return userNameCache.get(user);
        }

        const res = await webClient.users.info({user});
        const userName = res.ok ? res.user.name : user;

        // Cache the result for future calls
        userNameCache.set(user, userName);

        return userName;
    });
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

    try {
        const gptscriptClient = new GPTScript()
        const elements = []
        for (const [key, value] of data.entries()) {
            elements.push({
                name: `${key}`,
                description: '',
                contents: `${value}`
            })
        }
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `slack_history_${channelId}`,
            description: `chat history for channel "${channelId}"`
        })
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} messages`)
    } catch (e) {
        console.log('Failed to create dataset:', e)
    }
}
