import base64
import json
import os
import time
from collections.abc import Mapping
from typing import Any, AsyncIterable, Iterator, List, Optional

import requests
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse
from ollama import Client, Options
from openai.types.chat import ChatCompletionChunk
from openai.types.chat.chat_completion_chunk import Choice, ChoiceDelta, ChoiceDeltaToolCall, \
    ChoiceDeltaToolCallFunction

debug = os.environ.get('DEBUG', False) == "true"
uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")
ollama_host = os.environ.get("OBOT_OLLAMA_MODEL_PROVIDER_HOST", "127.0.0.1:11436")
ollama_client = Client(host=ollama_host)


def log(*args):
    if debug:
        print(*args)


app = FastAPI()

# system: str = """
# You are task oriented system.
# You receive input from a user, process the input from the given instructions, and then output the result.
# Your objective is to provide consistent and correct results.
# Call the provided tools as needed to complete the task.
# You do not need to explain the steps taken, only provide the result to the given instructions.
# You are referred to as a tool.
# You don't move to the next step until you have a result.
# """


@app.middleware("http")
async def log_body(request: Request, call_next):
    body = await request.body()
    log("REQUEST BODY: ", body)
    return await call_next(request)


@app.get("/")
@app.post("/")
async def get_root():
    return uri


@app.get("/v1/models")
async def list_models() -> JSONResponse:
    data: list[dict] = []
    models = ollama_client.list()
    for model in models['models']:
        # truncate nanoseconds to microseconds
        timestamp = model['modified_at']
        unix_timestamp = time.mktime(timestamp.timetuple())

        data.append({
            "id": model['model'],
            "object": "model",
            "created": int(unix_timestamp),
            "owned_by": "local",
        })
    return JSONResponse(content={"object": "list", "data": data})


@app.post("/v1/chat/completions")
async def chat_completions(request: Request):
    data = await request.body()
    data = json.loads(data)

    messages = data.get("messages", [])
    if messages:
        if messages[0]["role"] == "system":
            messages[0]["role"] = 'user'
    # messages.insert(0, {"role": "system", "content": system})

    messages = merge_consecutive_dicts_with_same_value(messages, "role")

    for index, message in enumerate(messages):
        if message["role"] == "assistant" and message.get('tool_calls', None) is not None:
            for tool_call in message["tool_calls"]:
                if tool_call["function"].get('arguments', None) is not None:
                    parsed = json.loads(tool_call["function"]["arguments"])
                    tool_call["function"]["arguments"] = marshal_top_level(parsed)

        text_content: str | None = None
        image_content = []
        if type(message.get('content', None)) is list:
            for content in message['content']:
                if content['type'] == 'text':
                    text_content = content['text']
                if content['type'] == 'image_url':
                    if content['image_url']['url'].startswith("data:"):
                        image_content.append(content['image_url']['url'])
                    else:
                        image = requests.get(content['image_url']['url'])
                        image.raise_for_status()
                        b64_image = base64.b64encode(image.content).decode('utf-8')
                        image_content.append(b64_image)
            messages[index]['content'] = text_content
            messages[index]['images'] = image_content

    try:
        resp: Mapping[str, Any] | Iterator[Mapping[str, Any]] = ollama_client.chat(
            model=data['model'],
            messages=messages,
            tools=data.get("tools", None),
            stream=False,  # @TODO Change to allow streaming once it is supported # data.get("resp", False),
            options=Options(
                temperature=data.get("temperature", None),
                top_p=data.get("top_p", None),
            )
        )
    except Exception as e:
        status_code = e.__dict__.pop("status_code", 500)
        return JSONResponse(content={"error": str(e)}, status_code=status_code)

    async def convert_stream(stream: Mapping[str, Any] | Iterator[Mapping[str, Any]]) -> AsyncIterable[str]:
        # @TODO: Implement streaming once it is supported
        completion: ChatCompletionChunk
        if type(resp) is Iterator:
            for chunk in stream:
                tool_calls: Optional[List[ChoiceDeltaToolCall]] = None
                if resp['message'].get('tool_calls', None) is not None:
                    tool_calls = []
                    for index, tool_call in enumerate(resp['message']['tool_calls']):
                        arguments: str | None = None
                        if tool_call['function'].get('arguments', None) is not None:
                            arguments = json.dumps(tool_call['function']['arguments'])
                        call = ChoiceDeltaToolCall(
                            index=index,
                            id=tool_call['function']['name'],
                            function=ChoiceDeltaToolCallFunction(
                                name=tool_call['function'].get('name', None),
                                arguments=arguments,
                            ),
                            type='function',
                        )
                        tool_calls.append(call)
                # @TODO map finish_reasons to the correct values for streaming
                finish_reason: str | None = None
                if chunk.get("done", None) is True:
                    finish_reason = chunk['done_reason']
                    if type(tool_calls) is not None:
                        finish_reason = 'tool_calls'

                completion = ChatCompletionChunk(
                    id='0',
                    choices=[Choice(
                        delta=ChoiceDelta(
                            content=chunk["message"].get('content', None),
                            role=chunk["message"].get('role', None),
                            tool_calls=tool_calls,
                        ),
                        # "stop", "length", "tool_calls", "content_filter", None
                        finish_reason=finish_reason,
                        index=0,
                    )],
                    created=0,
                    model=chunk['model'],
                    object="chat.completion.chunk",
                )
        else:
            tool_calls: Optional[List[ChoiceDeltaToolCall]] = None
            if resp['message'].get('tool_calls', None) is not None:
                tool_calls = []
                for index, tool_call in enumerate(resp['message']['tool_calls']):
                    arguments: str | None = None
                    if tool_call['function'].get('arguments', None) is not None:
                        arguments = json.dumps(tool_call['function']['arguments'])
                    call = ChoiceDeltaToolCall(
                        index=index,
                        id=tool_call['function']['name'],
                        function=ChoiceDeltaToolCallFunction(
                            name=tool_call['function'].get('name', None),
                            arguments=arguments,
                        ),
                        type='function',
                    )
                    tool_calls.append(call)

            finish_reason = 'stop'
            if type(tool_calls) is not None:
                finish_reason = 'tool_calls'

            completion = ChatCompletionChunk(
                id='0',
                choices=[Choice(
                    delta=ChoiceDelta(
                        content=resp["message"].get('content', None),
                        role=resp["message"].get('role', None),
                        tool_calls=tool_calls,
                    ),
                    finish_reason=finish_reason,
                    index=0,
                )],
                created=0,
                model=resp['model'],
                object="chat.completion.chunk",
            )
        transformed = completion.model_dump(mode="json", exclude_unset=True, exclude_none=True)
        for choice in transformed['choices']:
            if choice['delta'].get('tool_calls', None) is not None:
                for index, tool_call in enumerate(choice['delta']['tool_calls']):
                    tool_call["index"] = index

        log("CHUNK: ", json.dumps(transformed))
        yield "data: " + json.dumps(transformed) + "\n\n"

    return StreamingResponse(convert_stream(resp), media_type="application/x-ndjson")


def marshal_top_level(json_dict):
    top_level_map = {}
    for key, value in json_dict.items():
        if isinstance(value, (dict, list)):
            top_level_map[key] = json.dumps(value)
        else:
            top_level_map[key] = value
    return top_level_map


def merge_consecutive_dicts_with_same_value(list_of_dicts, key) -> list[dict]:
    merged_list = []
    index = 0
    while index < len(list_of_dicts):
        current_dict = list_of_dicts[index]
        value_to_match = current_dict.get(key)
        compared_index = index + 1
        while compared_index < len(list_of_dicts) and list_of_dicts[compared_index].get(key) == value_to_match:
            list_of_dicts[compared_index]["content"] = current_dict["content"] + "\n" + list_of_dicts[compared_index][
                "content"]
            current_dict.update(list_of_dicts[compared_index])
            compared_index += 1
        merged_list.append(current_dict)
        index = compared_index
    return merged_list


######
if __name__ == "__main__":
    import uvicorn
    import asyncio

    try:
        uvicorn.run("main:app", host="127.0.0.1", port=int(os.environ.get("PORT", "8000")), workers=4,
                    log_level="debug" if debug else "critical", access_log=debug)
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass