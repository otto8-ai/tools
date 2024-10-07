from youtube_transcript_api import YouTubeTranscriptApi
from youtube_transcript_api._errors import NoTranscriptAvailable, TranscriptsDisabled

from helpers import get_video_url

video_id = get_video_url()

try:
    transcript_list = YouTubeTranscriptApi.get_transcript(video_id)
    transcript_text = ' '.join(item['text'] for item in transcript_list)
    print(transcript_text)
except NoTranscriptAvailable:
    print(f"Error: No transcript was available for video {video_id}")
except TranscriptsDisabled:
    print(f"Error: Transcripts are disabled for video {video_id}")
except Exception as e:
    print(f"Error: {e}")
