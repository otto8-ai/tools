import json
import os
from typing import AsyncIterable

from fastapi import FastAPI, HTTPException, Request
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse, StreamingResponse
from openai import OpenAI
from openai._streaming import Stream
from openai.types import CreateEmbeddingResponse, ImagesResponse
from openai.types.chat import ChatCompletion, ChatCompletionChunk

api_key = os.environ.get("OTTO8_OPENAI_MODEL_PROVIDER_API_KEY", "")
debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"
uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")

client = OpenAI(api_key=api_key)


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
    try:
        models = client.models.list()
        return JSONResponse(content={"object":"list","data": [set_model_usage(m) for m in models.to_dict()["data"]]})
    except Exception as e:
        print(e)

def set_model_usage(model: dict) -> dict:
    if (model["id"].startswith("gpt-") or model["id"].startswith("ft:gpt-") or model["id"].startswith("o1-") or model["id"].startswith("ft:o1-")) and "-realtime-" not in model["id"]:
        model["metadata"] = {"usage":"llm"}
    elif model["id"].startswith("text-embedding") or model["id"].startswith("ft:text-embedding"):
        model["metadata"] = {"usage":"text-embedding"}
    elif model["id"].startswith("dalle-") or model["id"].startswith("ft:dalle-"):
        model["metadata"] = {"usage":"image-generation"}
    return model


@app.post("/v1/chat/completions")
async def chat_completions(request: Request):
    data = await request.body()
    data = json.loads(data)

    stream = data.get("stream", False)

    messages = data["messages"]
    messages.insert(0, {"content": system, "role": "system"})

    try:
        res: Stream[ChatCompletionChunk] | ChatCompletion = client.chat.completions.create(**data)
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


@app.post("/v1/embeddings")
async def embeddings(request: Request):
    data = json.loads(await request.body())

    try:
        res: CreateEmbeddingResponse = client.embeddings.create(**data)

        return JSONResponse(content=jsonable_encoder(res))
    except Exception as e:
        try:
            log("Error occurred: ", e.__dict__)
            error_code = e.status_code
            error_message = e.message
        except:
            error_code = 500
            error_message = str(e)
        raise HTTPException(status_code=error_code, detail=f"Error occurred: {error_message}")


@app.post("/v1/images/generations")
async def image_generation(request: Request):
    data = json.loads(await request.body())

    try:
        res: ImagesResponse = client.images.generate(**data)

        return JSONResponse(content=jsonable_encoder(res))
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
