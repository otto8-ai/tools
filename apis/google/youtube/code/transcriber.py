import os
import urllib.parse as urlparse
from urllib.parse import parse_qs

from youtube_transcript_api import YouTubeTranscriptApi

video_url = os.getenv('VIDEO_URL')
if video_url is None:
    raise ValueError("Error: video_url must be set")

parsed = urlparse.urlparse(video_url)
if parsed.netloc == 'www.youtube.com':
    video_id = parse_qs(parsed.query)['v'][0]
elif parsed.netloc == 'youtu.be':
    video_id = parsed.path.strip('/').split('/')[0]
else:
    raise ValueError(f"Invalid YouTube URL: {video_url}")

try:
    transcript_list = YouTubeTranscriptApi.get_transcript(video_id)
    transcript_text = ' '.join(item['text'] for item in transcript_list)
    print(transcript_text)
except Exception as e:
    print(f"Error: {e}")