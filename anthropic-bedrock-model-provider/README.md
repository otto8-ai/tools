# Anthropic Bedrock Provider

## Usage Example

```
gptscript --default-model='anthropic.claude-3-opus-20240229-v1:0 from github.com/gptscript-ai/claude3-bedrock-provider' examples/helloworld.gpt
```

The native Anthropic provider can be found [here](https://github.com/obot-platform/tools/anthropic-model-provider)

## Development

* You need to have AWS credentials and a region configured in your environment.

Run using the following commands

```
python -m venv .venv
source ./.venv/bin/activate
pip install --upgrade -r requirements.txt
./run.sh
```

```
export OPENAI_BASE_URL=http://127.0.0.1:8000/v1
export GPTSCRIPT_DEBUG=true

gptscript --default-model=anthropic.claude-3-opus-20240229-v1:0 examples/bob.gpt
```
