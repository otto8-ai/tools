import {GPTScript} from "@gptscript-ai/gptscript";
import {Mutex} from "async-mutex";

export async function listChannels(webClient) {
    const [publicChannels, privateChannels] = await Promise.all([
        webClient.conversations.list({limit: 50, types: 'public_channel'})
            .then(result => result.channels),
        webClient.conversations.list({limit: 50, types: 'private_channel'})
            .then(result => result.channels)
    ]);

    const channels = [...publicChannels, ...privateChannels] ?? []
    if (channels.length < 1) {
        console.log('No channels found')
        return
    }

    if (channels.length <= 10) {
        console.log(channels.map(channelToString).join(', '))
        return
    }

    try {
        const gptscriptClient = new GPTScript()
        const dataset = await gptscriptClient.createDataset(
            process.env.GPTSCRIPT_WORKSPACE_ID,
            'slack_channels',
            'list of slack channels'
        )

        await Promise.all(
            channels.map(channel =>
                gptscriptClient.addDatasetElement(
                    process.env.GPTSCRIPT_WORKSPACE_ID,
                    dataset.id,
                    channel.name,
                    channel.purpose.value || '',
                    channelToString(channel)
                )
            )
        );

        console.log(`Created dataset with ID ${dataset.id} with ${channels.length} channels`)
    } catch (e) {
        console.log(`Error creating dataset: ${e}`)
    }

}

export async function searchChannels(webClient, query) {
    const [publicChannels, privateChannels] = await Promise.all([
        webClient.conversations.list({limit: 50, types: 'public_channel'})
            .then(result => result.channels.filter(channel => channel.name.includes(query))),

        webClient.conversations.list({limit: 50, types: 'private_channel'})
            .then(result => result.channels.filter(channel => channel.name.includes(query)))
    ]);

    const channels = [...publicChannels, ...privateChannels] ?? [];

    if (channels.length < 1) {
        console.log('No channels found')
        return
    }

    if (channels.length <= 10) {
        console.log(
            channels
                .filter(channel => channel.name.includes(query))
                .map(channelToString)
                .join(',')
        )
        return
    }

    try {
        const gptscriptClient = new GPTScript()
        const dataset = await gptscriptClient.createDataset(
            process.env.GPTSCRIPT_WORKSPACE_ID,
            `${query}_slack_channels`,
            `list of slack channels matching search query "${query}"`
        )

        await Promise.all(
            channels.map(channel =>
                gptscriptClient.addDatasetElement(
                    process.env.GPTSCRIPT_WORKSPACE_ID,
                    dataset.id,
                    channel.name,
                    channel.purpose.value || '',
                    channelToString(channel)
                )
            )
        );

        console.log(`Created dataset with ID ${dataset.id} with ${channels.length} channels`)
    } catch (e) {
        console.log(`Error creating dataset: ${e}`)
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
    const history = await webClient.conversations.history({
        channel: channelId,
        limit: limit,
        oldest: oldest.toString(),
        latest: latest.toString()
    })
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
        console.log(await messageToString(webClient, reply))
    }
}


export async function search(webClient, query) {
    const result = await webClient.search.messages({query});

    if (!result.ok) {
        console.log(`Failed to search messages: ${result.error}`);
        process.exit(1);
    }

    if (result.messages.matches.length < 1) {
        console.log('No messages found');
        return;
    }

    const messages = result.messages.matches;

    // If more than 10 messages are found, add them to a dataset
    if (messages.length > 10) {
        try {
            const gptscriptClient = new GPTScript();
            const dataset = await gptscriptClient.createDataset(
                process.env.GPTSCRIPT_WORKSPACE_ID,
                `slack_search_${query}`,
                `search results for query "${query}"`
            );

            await Promise.all(
                messages.map(async (message) => {
                    return gptscriptClient.addDatasetElement(
                        process.env.GPTSCRIPT_WORKSPACE_ID,
                        dataset.id,
                        `${message.iid}_${message.ts}`,
                        "",
                        await messageToString(webClient, message)
                    )
                })
            );

            console.log(`Created dataset with ID ${dataset.id} with ${messages.length} search results`);
            return;
        } catch (e) {
            console.log(`Error creating dataset: ${e}`);
        }
    }

    // Process and print each message concurrently if not added to a dataset
    await Promise.all(
        messages.map(async (message) => {
            const messageText = await messageToString(webClient, message);
            console.log(messageText);
        })
    );
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
        } catch (e) {
        } // Ignore errors if we got any. We'll just print the results below.
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
        } catch (e) {
        } // Ignore errors if we got any. We'll just print the results below.
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
    }

    for (const reply of replies.messages) {
        console.log(await messageToString(webClient, reply))
    }
}


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
    return [
        `${user.name}`,
        `  ID: ${user.id}`,
        `  Full name: ${user.profile.real_name}`,
        user.deleted ? '  Account deleted: true' : ''
    ].filter(Boolean).join('\n');
}

async function messageToString(webClient, message) {
    const time = new Date(parseFloat(message.ts) * 1000).toLocaleString();
    let userName = message.user;

    try {
        userName = await getUserName(webClient, message.user);
    } catch (e) {
    }

    const userMentions = message.text.match(/<@U[A-Z0-9]+>/g) ?? [];
    for (const mention of userMentions) {
        const userId = mention.substring(2, mention.length - 1);
        try {
            const mentionUserName = await getUserName(webClient, userId);
            message.text = message.text.replace(mention, `@${mentionUserName}`);
        } catch (e) {
        }
    }

    return [
        `${time}: ${userName}: ${message.text}`,
        `  message ID: ${message.ts}`,
        message.blocks?.length ? `  message blocks: ${JSON.stringify(message.blocks)}` : '',
        message.attachments?.length ? `  message attachments: ${JSON.stringify(message.attachments)}` : ''
    ].filter(Boolean).join('\n');
}

function channelToString(channel) {
    return `${channel.name} (ID: ${channel.id})${channel.is_archived ? ' (archived)' : ''}`;
}

// Helper function to fetch and format a single message and its replies
async function fetchMessageWithReplies(webClient, channelId, message) {
    const messageGroup = [await messageToString(webClient, message)]; // Main message

    if (message.reply_count > 0) {
        // Fetch replies for this message
        const replies = await webClient.conversations.replies({
            channel: channelId,
            ts: message.ts,
            limit: 3
        });

        // Add a summary line for the thread
        messageGroup.push(`  thread ID ${threadID(message)} - ${message.reply_count} ${replyString(message.reply_count)}:`);

        // Collect formatted replies, excluding the main message if it’s included
        const replyTexts = await Promise.all(
            replies.messages
                .filter(reply => reply.ts !== message.ts)
                .map(reply => messageToString(webClient, reply))
        );

        messageGroup.push(...replyTexts);

        if (replies.has_more) {
            messageGroup.push('  More replies exist');
        }
    }

    return messageGroup;
}

// Main function to print the channel history, maintaining order and logging only once
export async function printHistory(webClient, channelId, history) {
    // Fetch each message with its replies in parallel
    const fetchPromises = history.messages.map(message => fetchMessageWithReplies(webClient, channelId, message));

    // Resolve all fetches, preserving the order of main messages and replies
    const allMessagesData = await Promise.all(fetchPromises);

    // Accumulate all message groups into a single output string
    const output = allMessagesData.map(messageGroup => messageGroup.join('\n')).join('\n\n');

    // Log the entire output at once
    console.log(output);
}
