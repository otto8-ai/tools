import io
import sys
import os

from googleapiclient.http import MediaIoBaseDownload

from auth import client
from id import extract_file_id

def main():
    try:
        doc_ref = os.getenv('DOC_REF')
        if not doc_ref:
            raise ValueError('DOC_REF environment variable is missing or empty')

        file_id = extract_file_id(doc_ref)
        service = client('docs', 'v1')
        document = service.documents().get(documentId=file_id).execute()

        print(convert_to_markdown(document))

    except Exception as err:
        sys.stderr.write(err)
        sys.exit(1)

def convert_to_markdown(document):
    md_text = ""
    for element in document.get('body', {}).get('content', []):
        if 'paragraph' in element:
            for part in element['paragraph']['elements']:
                text_run = part.get('textRun')
                if text_run:
                    md_text += text_run['content']
            md_text += "\n\n"  # Separate paragraphs with extra newlines
        elif 'table' in element:
            md_text += parse_table(element['table'])
            md_text += "\n\n"  # Extra newline after a table
    return md_text

def parse_table(table):
    md_table = ""
    for row in table.get('tableRows', []):
        row_text = "|"
        for cell in row.get('tableCells', []):
            cell_text = ""
            for content in cell.get('content', []):
                if 'paragraph' in content:
                    for element in content['paragraph']['elements']:
                        text_run = element.get('textRun')
                        if text_run:
                            cell_text += text_run['content']
            row_text += f" {cell_text.strip()} |"
        md_table += row_text + "\n"
    return md_table

if __name__ == "__main__":
    main()
