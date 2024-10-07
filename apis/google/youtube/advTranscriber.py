import glob
import os

import yt_dlp
from openai import OpenAI

from helpers import get_video_url

video_id = get_video_url()

download_dir = os.getenv('GPTSCRIPT_WORKSPACE_DIR', '/tmp')
filename = 'audio'
output = f'{download_dir}/{filename}'
ydl_opts = {
    'format': 'worstaudio[ext=m4a]/worstaudio[ext=webm]/bestaudio[abr<=64k][ext=ogg]/bestaudio[abr<=64k]',
    'outtmpl': f'{output}.%(ext)s',
}

with yt_dlp.YoutubeDL(ydl_opts) as ydl:
    error_code = ydl.download(video_id)

files = glob.glob(f'{output}.*')
if not files:
    raise ('error: no downloaded files found, cannot continue')

size_in_bytes = os.path.getsize(files[0])
size_in_mb = round(size_in_bytes / (1024 * 1024), 2)
if size_in_mb >= 25.00:
    raise ('error: file size is too large, cannot continue')

client = OpenAI()
try:
    audio_file = open(files[0], "rb")
    transcript_text = client.audio.transcriptions.create(
        model="whisper-1",
        file=audio_file
    )
    print(transcript_text.text)
    audio_file.close()
    os.remove(files[0])
except Exception as e:
    raise (f"Error: {e}")
