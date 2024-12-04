import asyncio
import os
from helpers import client, list_drafts

async def main():
    max_results = os.getenv('MAX_RESULTS', '100')
    if max_results is not None:
        max_results = int(max_results)
    
    service = client('gmail', 'v1')

    await list_drafts(service, max_results)

if __name__ == "__main__":
    asyncio.run(main())
