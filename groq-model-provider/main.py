import base64
import json
import os
import time
from typing import AsyncIterable

import requests
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse
from openai import OpenAI

debug = os.environ.get("DEBUG", False) == "true"
uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")

# Initialize Groq client
client = OpenAI(
    base_url="https://api.groq.com/openai/v1",
    api_key=os.environ.get("OBOT_GROQ_MODEL_PROVIDER_API_KEY", ""),
)


def log(*args):
    if debug:
        print(*args)


app = FastAPI()


@app.middleware("http")
async def log_body(request: Request, call_next):
    if debug:
        body = await request.body()
        log("REQUEST BODY: ", body.decode())
    return await call_next(request)


@app.get("/")
@app.post("/")
async def get_root():
    return uri


@app.get("/v1/models")
async def list_models() -> JSONResponse:
    try:
        models = client.models.list()
        data = [
            {
                "id": model.id,
                "object": "model",
                "created": int(time.time()),
                "owned_by": "groq",
            }
            for model in models.data
        ]
        return JSONResponse(content={"object": "list", "data": data})
    except Exception as e:
        return JSONResponse(content={"error": str(e)}, status_code=500)


@app.post("/v1/chat/completions")
async def chat_completions(request: Request):
    data = await request.json()

    messages = data.get("messages", [])
    messages = merge_consecutive_dicts_with_same_value(messages, "role")

    # Handle image content in messages
    for index, message in enumerate(messages):
        text_content = None
        image_content = []
        if isinstance(message.get("content", None), list):
            for content in message["content"]:
                if content["type"] == "text":
                    text_content = content["text"]
                if content["type"] == "image_url":
                    if content["image_url"]["url"].startswith("data:"):
                        image_content.append(content["image_url"]["url"])
                    else:
                        image = requests.get(content["image_url"]["url"])
                        image.raise_for_status()
                        b64_image = base64.b64encode(image.content).decode("utf-8")
                        image_content.append(b64_image)
            messages[index]["content"] = text_content
            if image_content:
                messages[index]["images"] = image_content

    try:
        # Always use streaming for Groq
        response = client.chat.completions.create(
            model=data["model"],
            messages=messages,
            tools=data.get("tools", None),
            stream=True,  # Always True for Groq
            temperature=data.get("temperature", None),
            top_p=data.get("top_p", None),
        )

        return StreamingResponse(
            convert_stream(response), media_type="application/x-ndjson"
        )

    except Exception as e:
        return JSONResponse(content={"error": str(e)}, status_code=500)


async def convert_stream(stream) -> AsyncIterable[str]:
    for chunk in stream:
        transformed = chunk.model_dump(
            mode="json", exclude_unset=True, exclude_none=True
        )
        if "choices" in transformed:
            for choice in transformed["choices"]:
                if choice.get("delta", {}).get("tool_calls"):
                    for index, tool_call in enumerate(choice["delta"]["tool_calls"]):
                        tool_call["index"] = index

        log("CHUNK: ", json.dumps(transformed))
        yield "data: " + json.dumps(transformed) + "\n\n"


def merge_consecutive_dicts_with_same_value(list_of_dicts, key) -> list[dict]:
    merged_list = []
    index = 0
    while index < len(list_of_dicts):
        current_dict = list_of_dicts[index]
        value_to_match = current_dict.get(key)
        compared_index = index + 1
        while (
            compared_index < len(list_of_dicts)
            and list_of_dicts[compared_index].get(key) == value_to_match
        ):
            list_of_dicts[compared_index]["content"] = (
                current_dict["content"]
                + "\n"
                + list_of_dicts[compared_index]["content"]
            )
            current_dict.update(list_of_dicts[compared_index])
            compared_index += 1
        merged_list.append(current_dict)
        index = compared_index
    return merged_list


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
