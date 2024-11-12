import json
import os

import claude3_provider_common
from anthropic import AsyncAnthropic
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse

debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"
client = AsyncAnthropic(api_key=os.environ.get("OTTO8_ANTHROPIC_MODEL_PROVIDER_API_KEY", ""))
app = FastAPI()
uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")


def log(*args):
    if debug:
        print(*args)


@app.middleware("http")
async def log_body(request: Request, call_next):
    body = await request.body()
    log("HTTP REQUEST BODY: ", body)
    return await call_next(request)


@app.post("/")
async def post_root():
    return uri


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
        uvicorn.run("main:app", host="127.0.0.1", port=int(os.environ.get("PORT", "8000")),
                log_level="debug" if debug else "critical", access_log=debug)
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass
