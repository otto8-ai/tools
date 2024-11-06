import json
import os
from typing import AsyncIterable

from fastapi import FastAPI, HTTPException, Request
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse, StreamingResponse
from openai._streaming import Stream
from openai._types import NOT_GIVEN
from openai.types.chat import ChatCompletion, ChatCompletionChunk

import helpers

debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"
uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")


def log(*args):
    if debug:
        print(*args)


app = FastAPI()

system: str = """
You are task oriented system.
You receive input from a user, process the input from the given instructions, and then output the result.
Your objective is to provide consistent and correct results.
You do not need to explain the steps taken, only provide the result to the given instructions.
You are referred to as a tool.
You don't move to the next step until you have a result.
"""


@app.middleware("http")
async def log_body(request: Request, call_next):
    body = await request.body()
    log("REQUEST BODY: ", body)
    return await call_next(request)


@app.get("/")
@app.post("/")
async def get_root():
    return uri


# Only needed when running standalone. With GPTScript, the `id` returned by this endpoint must match the model (deployment) you are passing in.
@app.get("/v1/models")
async def list_models() -> JSONResponse:
    return JSONResponse(content={"data": [{"id": "gpt-4", "name": "Your model"}]})


@app.post("/v1/chat/completions")
async def chat_completions(request: Request):
    data = await request.body()
    data = json.loads(data)

    tools = data.get("tools", NOT_GIVEN)

    if tools is not NOT_GIVEN:
        tool_choice = 'auto'
    else:
        tool_choice = NOT_GIVEN

    temperature = data.get("temperature", NOT_GIVEN)
    if temperature is not NOT_GIVEN:
        temperature = float(temperature)

    stream = data.get("stream", False)

    messages = data["messages"]
    messages.insert(0, {"content": system, "role": "system"})

    config = await helpers.get_azure_config(data["model"])
    if config == None:
        raise HTTPException(status_code=400,
                            detail="Azure config not found. Please ensure you have configured the environment variables correctly.")

    client = helpers.client(
        endpoint=config.endpoint,
        deployment_name=config.deployment_name,
        api_key=config.api_key
    )
    try:
        res: Stream[ChatCompletionChunk] | ChatCompletion = client.chat.completions.create(model=data["model"],
                                                                                           messages=messages,
                                                                                           tools=tools,
                                                                                           tool_choice=tool_choice,
                                                                                           temperature=temperature,
                                                                                           stream=stream
                                                                                           )
        if not stream:
            return JSONResponse(content=jsonable_encoder(res))

        return StreamingResponse(convert_stream(res), media_type="application/x-ndjson")
    except Exception as e:
        try:
            log("Error occurred: ", e.__dict__)
            error_code = e.status_code
            error_message = e.message
        except:
            error_code = 500
            error_message = str(e)
        raise HTTPException(status_code=error_code, detail=f"Error occurred: {error_message}")


async def convert_stream(stream: Stream[ChatCompletionChunk]) -> AsyncIterable[str]:
    for chunk in stream:
        log("CHUNK: ", chunk.model_dump_json())
        yield "data: " + str(chunk.model_dump_json()) + "\n\n"


if __name__ == "__main__":
    import uvicorn
    import asyncio

    try:
        uvicorn.run("main:app", host="127.0.0.1", port=int(os.environ.get("PORT", "8000")),
                    log_level="debug" if debug else "critical", access_log=debug)
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass