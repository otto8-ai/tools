import json
import os

from llama_index.core import SimpleDirectoryReader

docs = SimpleDirectoryReader(input_files=[os.getenv("INPUT")]).load_data()

texts = []

for doc in docs:
    if len(doc.get_text()) == 0:
        continue
    texts.append(f"!metadata {json.dumps(doc.metadata)}\n{doc.get_text()}")

with open(os.getenv("OUTPUT"), "w") as f:
    f.write("\n---docbreak---\n".join(texts))
