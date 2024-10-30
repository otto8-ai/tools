# GPTScript Knowledge Tool

Standalone Knowledge Tool to be used with GPTScript and GPTStudio

**[gptscript-ai.github.io/knowledge](https://gptscript-ai.github.io/knowledge)**

## Build

Requires Go 1.22+

```bash
make build
```

## Run

The knowledge tool can run in two modes: server and client, where client can be standalone or referring to a remote server.

You can find a full gptscript-generated documentation in the [CLI documentation](./docs/cli.md).


### Client - Standalone

```bash
knowledge create-dataset foobar
knowledge ingest -d foobar README.md
knowledge retrieve -d foobar "Which filetypes are supported?"
knowledge delete-dataset foobar
```

### Server & Client - Server Mode

**WARNING** The server mode is not fully implemented and currently lacking some features. You're well advised to use the standalone client mode.

<details>

<summary>Server</summary>

```bash
knowledge server
```

```bash
export KNOW_SERVER_URL=http://localhost:8000/v1
knowledge create-dataset foobar
knowledge ingest -d foobar README.md
knowledge retrieve -d foobar "Which filetypes are supported?"
knowledge delete-dataset foobar
```

</details>


## Supported File Types

- `.pdf`
- `.html`
- `.md`
- `.txt`
- `.docx`
- `.odt`
- `.rtf`
- `.csv`
- `.ipynb`
- `.json`

## OpenAPI / Swagger

The API is documented using OpenAPI 2.0 (Swagger), automatically generated using [`swaggo/swag`](https://github.com/swaggo/swag) (`make openapi`).

## GPTScript Examples

Note: The examples in the `examples/` directory expect the `knowledge` binary to be in your `$PATH`.

### Run

```bash
gptscript examples/client.gpt
```

## Choosing an Embedding Model

**! Important !**

The same embedding model must be used for both ingestion and retrieval.

Checkout the MTEB Leaderboard: https://huggingface.co/spaces/mteb/leaderboard