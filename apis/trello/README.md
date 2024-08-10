# Trello API Tool

This is a set of GPTScript tools to interact with the Trello API.

## Prerequisite

You need generate a Trello API token [here](https://trello.com/1/authorize?expiration=never&scope=read,write,account&response_type=token&key=aabf1a5f6af0a2e5c4c3807b4d3ccbc8). You will be automatically prompted for this information when you are running the tool.

## Quick Start

Create a new file called `agent.gpt` with this as the contents:

```
Tools: github.com/gptscript-ai/tools/apis/trello
Chat: true

Please help me with my Trello workspace.
```

Then, run it with `gptscript agent.gpt`. You'll be prompted to enter your API key; once you provide it, you'll be able to chat with the large language model and have it perform actions for you in Trello.

## Tools

These are some of the tools available to the LLM when you use this tool set:

- **post-cards**: Create new cards on a Trello board.
- **delete-cards-id**: Delete a card by its ID.
- **get-boards-id-cards**: Retrieve all cards on a specific board.
- **get-organizations-id-boards**: Retrieve all boards in a specific organization.
- **get-members-id-organizations**: Retrieve all organizations a member belongs to.
