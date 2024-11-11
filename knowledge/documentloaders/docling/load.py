import os

from docling_core.types.doc import ImageRefMode, PictureItem, TableItem
from docling.datamodel.base_models import FigureElement, InputFormat, Table
from docling.datamodel.pipeline_options import PdfPipelineOptions
from docling.document_converter import DocumentConverter, PdfFormatOption

source = os.getenv("INPUT")


IMAGE_RESOLUTION_SCALE = 2.0

pipeline_options = PdfPipelineOptions()
pipeline_options.images_scale = IMAGE_RESOLUTION_SCALE
pipeline_options.generate_picture_images = True

converter = DocumentConverter(
    format_options={
        InputFormat.PDF: PdfFormatOption(pipeline_options=pipeline_options)
    }
)
result = converter.convert(source)

# TODO: send embedded images to VLM for captioning


with open(os.getenv("OUTPUT"), "w") as f:
    f.write(result.document.export_to_markdown(image_mode=ImageRefMode.EMBEDDED))