import json
import os
import sys

import boto3
import claude3_provider_common
from anthropic import AsyncAnthropicBedrock
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse

debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"
def log(*args):
    if debug:
        print(*args)


os.environ["AWS_ACCESS_KEY_ID"] = os.environ.get("OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_ACCESS_KEY_ID")
os.environ["AWS_SECRET_ACCESS_KEY"] = os.environ.get("OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SECRET_ACCESS_KEY")
os.environ["AWS_SESSION_TOKEN"] = os.environ.get("OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SESSION_TOKEN")
os.environ["AWS_REGION"] = os.environ["AWS_DEFAULT_REGION"] = os.environ.get("OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_REGION")

# Check if any is empty
if not all([os.environ["OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_ACCESS_KEY_ID"], os.environ["OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SECRET_ACCESS_KEY"], os.environ["OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SESSION_TOKEN"], os.environ["OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_REGION"]]):
    print("Please set OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_ACCESS_KEY_ID, OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SECRET_ACCESS_KEY, OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SESSION_TOKEN, OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_REGION", file=sys.stderr)
    sys.exit(1)


# Check authentication
try:
    client = boto3.client("sts")
    response = client.get_caller_identity()
except Exception as e:
    print("Please authenticate with AWS - ", e, file=sys.stderr)
    sys.exit(1)



# Setup Client
client = AsyncAnthropicBedrock()

app = FastAPI()

uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")


@app.middleware("http")
async def log_body(request: Request, call_next):
    body = await request.body()
    log("HTTP REQUEST BODY: ", body)
    return await call_next(request)


@app.post("/")
@app.get("/")
async def get_root():
    return uri


@app.get("/v1/models")
async def list_models() -> JSONResponse:
    return await claude3_provider_common.list_models(client)


@app.post("/v1/chat/completions")
async def completions(request: Request) -> StreamingResponse:
    data = await request.body()
    input = json.loads(data)
    return await claude3_provider_common.completions(client, input)


if __name__ == "__main__":
    import uvicorn
    import asyncio

    try:
        uvicorn.run(
            "main:app",
            host="127.0.0.1",
            port=int(os.environ.get("PORT", "8000")),
            workers=4,
            log_level="debug" if debug else "critical",
            access_log=debug,
        )
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass
