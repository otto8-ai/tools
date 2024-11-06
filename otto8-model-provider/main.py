import os
import time
from typing import Any

import uvicorn
import asyncio
import httpx
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse

debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"
otto_url = os.environ.get("OTTO8_URL", "http://localhost:8080")
app = FastAPI()


@app.middleware("http")
async def log_body(request: Request, call_next):
    if debug:
        body = await request.body()
        print("HTTP REQUEST BODY: ", body)
    return await call_next(request)


@app.get("/")
@app.post("/")
async def root():
    return 'ok'


@app.get("/v1/models")
async def list_models() -> JSONResponse:
    # Collect all the LLM providers
    resp = httpx.get(f"{otto_url}/api/models")
    if resp.status_code != 200:
        return JSONResponse({"data": [], "error": resp.text}, status_code=resp.status_code)

    return JSONResponse({"data": [{"id": model["id"]} for model in resp.json()["items"]]}, status_code=200)


@app.post("/v1/chat/completions")
async def completions(request: Request) -> StreamingResponse:
    resp = _stream_chat_completion(await request.json(), get_api_key(request))
    status_code = 0
    async for code in resp:
        status_code = code
        break

    return StreamingResponse(
        resp,
        media_type='text/event-stream',
        status_code=status_code,
    )


async def _stream_chat_completion(content: Any, api_key: str):
    async with httpx.AsyncClient(timeout=httpx.Timeout(30 * 60.0)) as client:
        async with client.stream(
                "POST",
                f"{otto_url}/api/llm-proxy/chat/completions",
                json=content,
                headers={
                    "Authorization": f"Bearer {api_key}",
                    "Accept": "text/event-stream",
                    "Accept-Encoding": "gzip",
                },
        ) as resp:
            yield resp.status_code
            async for chunk in resp.aiter_bytes():
                yield chunk

def get_api_key(request: Request) -> str:
    env_header =request.headers.get("X-GPTScript-Env")
    if env_header is None:
        return ""

    for env in env_header.split(","):
        if env.startswith("GPTSCRIPT_MODEL_PROVIDER_TOKEN="):
            return env[len("GPTSCRIPT_MODEL_PROVIDER_TOKEN="):].strip()

    return ""


if __name__ == "__main__":
    try:
        config = uvicorn.Config("main:app", host="127.0.0.1", port=int(os.environ.get("PORT", "8000")),
                                log_level="debug" if debug else "critical", access_log=debug)
        server = uvicorn.Server(config)

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        cor = loop.create_task(server.serve())
        loop.run_until_complete(cor)
    except (KeyboardInterrupt, asyncio.CancelledError):
        loop.run_until_complete(server.shutdown())
        cor.cancel()
        cor.exception()
