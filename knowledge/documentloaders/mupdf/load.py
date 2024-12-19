import os
import pymupdf4llm
import pathlib

md_text = pymupdf4llm.to_markdown(os.getenv("INPUT"))

pathlib.Path(os.getenv("OUTPUT")).write_bytes(md_text.encode())
