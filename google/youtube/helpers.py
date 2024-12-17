import os
from urllib.parse import urlparse, parse_qs
from textwrap import dedent
from gptscript import GPTScript
from gptscript.tool import ToolDef
from gptscript.datasets import DatasetElement
from pydantic import BaseModel


def get_video_id() -> str:
    video_url = os.getenv('VIDEO_URL')
    if video_url is None:
        raise ValueError('Error: video_url must be set')

    parsed = urlparse(video_url)
    netloc = parsed.netloc.replace('www.', '')

    if netloc == 'youtube.com':
        if '/watch' in parsed.path:
            # Extract video ID from watch URL query parameter
            video_id = parse_qs(parsed.query).get('v')
            if not video_id:
                raise ValueError(f'No video ID found in URL: {video_url}')
            video_id = video_id[0]
        elif '/live' in parsed.path:
            # Extract video ID from live URL path segment, ensuring path ends with video ID
            path_segments = parsed.path.strip('/').split('/')
            if len(path_segments) < 2 or path_segments[0] != 'live':
                raise ValueError(f'Invalid YouTube live URL: {video_url}')
            video_id = path_segments[-1].split('?')[0]
        else:
            raise ValueError(f'Invalid YouTube URL path: {video_url}')
    elif netloc == 'youtu.be':
        # Extract video ID from shortened youtu.be URL path
        video_id = parsed.path.lstrip('/')
    else:
        raise ValueError(f'Invalid YouTube URL: {video_url}')

    return video_id


class Transcript(BaseModel):
    summary: str
    text: str


gptscript_client = GPTScript()


async def get_clean_transcript(text: str) -> Transcript:
    model = os.getenv('MODEL', os.getenv('OBOT_DEFAULT_LLM_MODEL', 'gpt-4o'))
    if text is None or text == '':
        raise ValueError('Error: no transcript text provided')

    tool = ToolDef(
        name='clean_transcript',
        modelName=model,
        jsonResponse=True,
        instructions=dedent(f'''
            Clean up the TEXT below by adding punctuation and correcting all misspellings and errors.
            Do not truncate or reword any of the TEXT, just add punctuation and correct mistakes.
            Afterwards, write a two sentence summary of the transcript and return an object containing the clean text and summary with the following JSON format:

            {{
                "summary": "<two sentence summary>",
                "text": "<clean text>"
            }}

            ===TEXT START===
            {text}
            ===TEXT END===

            Return only the final JSON object, nothing else.
        '''),
    )

    try:
        run = gptscript_client.evaluate(tool)
        transcript = Transcript.model_validate_json(await run.text())
        gptscript_client.close()
    except Exception as e:
        raise Exception(f'Error cleaning transcript text: {e}')

    return transcript


async def create_transcript_dataset(video_url: str, transcript: Transcript):
    dataset_id = await gptscript_client.add_dataset_elements(
        [
            DatasetElement(
                name='transcript',
                description=transcript.summary,
                contents=transcript.text,
            )
        ],
        name=f'transcript_{video_url}',
        description=f'Transcript of {video_url}',
    )

    print(f'Created dataset with ID {dataset_id} and 1 transcript')
