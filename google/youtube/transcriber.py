from youtube_transcript_api import YouTubeTranscriptApi
from youtube_transcript_api._errors import NoTranscriptAvailable, TranscriptsDisabled

from helpers import get_video_id, get_clean_transcript, create_transcript_dataset

async def transcribe_video():
    try:
        video_id = get_video_id()
        transcript_list = YouTubeTranscriptApi.get_transcript(video_id)
        text = ' '.join(item['text'] for item in transcript_list)
        transcript = await get_clean_transcript(text)
        await create_transcript_dataset(video_id, transcript)
    except NoTranscriptAvailable:
        print(f'Error: No transcript was available for video')
    except TranscriptsDisabled:
        print('Error: Transcripts are disabled for video')
    except Exception as e:
        print(f'Error: {e}')

import asyncio
asyncio.run(transcribe_video())
