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

        service = client('drive', 'v3')

        request = service.files().export_media(
            fileId=file_id,
            mimeType='text/markdown'
        )
        file = io.BytesIO()
        downloader = MediaIoBaseDownload(file, request)
        done = False

        while not done:
            _, done = downloader.next_chunk()

        print(file.getvalue().decode('utf-8'))

    except Exception as err:
        sys.stderr.write(err)
        sys.exit(1)


if __name__ == "__main__":
    main()
