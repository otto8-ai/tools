import json
import os

from fastapi import FastAPI, Request, HTTPException
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from voyageai import AsyncClient

debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"
client = AsyncClient(api_key=os.environ.get("OBOT_VOYAGE_MODEL_PROVIDER_API_KEY", ""))
app = FastAPI()
uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")

voyage_models = [
    {"id": "voyage-3","metadata":{"usage":"text-embedding"},"object":"model"},
    {"id": "voyage-3-lite","metadata":{"usage":"text-embedding"},"object":"model"},
    {"id": "voyage-finance-2","metadata":{"usage":"text-embedding"},"object":"model"},
    {"id": "voyage-multilingual-2","metadata":{"usage":"text-embedding"},"object":"model"},
    {"id": "voyage-law-2","metadata":{"usage":"text-embedding"},"object":"model"},
    {"id": "voyage-code-2","metadata":{"usage":"text-embedding"},"object":"model"},
]


def log(*args):
    if debug:
        print(*args)


@app.middleware("http")
async def log_body(request: Request, call_next):
    body = await request.body()
    log("HTTP REQUEST BODY: ", body)
    return await call_next(request)


@app.get("/")
@app.post("/")
async def root():
    return uri


@app.get("/v1/models")
async def list_models() -> JSONResponse:
    return JSONResponse({"object":"list","data": voyage_models}, status_code=200)


@app.post("/v1/embeddings")
async def embeddings(request: Request) -> JSONResponse:
    try:
        data = json.loads(await request.body())
        resp = await client.embed([data["input"]], model=data["model"])
        return JSONResponse(content=jsonable_encoder({"data": [{"embedding": e} for e in resp.embeddings]}))
    except Exception as e:
        try:
            log("Error occurred: ", e.__dict__)
            error_code = e.status_code
            error_message = e.message
        except:
            error_code = 500
            error_message = str(e)
        raise HTTPException(status_code=error_code, detail=f"Error occurred: {error_message}")


if __name__ == "__main__":
    import uvicorn
    import asyncio

    try:
        uvicorn.run("main:app", host="127.0.0.1", port=int(os.environ.get("PORT", "8000")),
                    log_level="debug" if debug else "critical", access_log=debug)
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass
