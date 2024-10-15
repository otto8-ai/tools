import asyncio
import os

from googleapiclient.errors import HttpError
from gptscript import GPTScript

from auth import client


async def main():
    service = client('drive', 'v3')
    try:
        # Query Google Drive for all Google Sheets files
        query = "mimeType='application/vnd.google-apps.spreadsheet'"
        page_token = None
        sheets = dict() # mapping of sheet name to sheet ID
        while True:
            results = service.files().list(q=query,
                                           pageSize=10,
                                           fields="nextPageToken, files(id, name)",
                                           pageToken=page_token,
                                           ).execute()

            items = results.get('files', [])
            if not items:
                print('No spreadsheets found.')
                break

            for item in items:
                sheets[item['name']] = item['id']

            page_token = results.get('nextPageToken')
            if not page_token:
                break

        if len(sheets) > 10:
            gptscript_client = GPTScript()
            dataset = await gptscript_client.create_dataset(
                os.getenv("GPTSCRIPT_WORKSPACE_DIR"),
                "google_sheets_list",
                "list of Google Sheets in Google Drive"
            )

            for sheet_name, sheet_id in sheets.items():
                await gptscript_client.add_dataset_element(
                    os.getenv("GPTSCRIPT_WORKSPACE_DIR"),
                    dataset.id,
                    sheet_name,
                    f"ID: {sheet_id}",
                    "sheet ID"
                )

            print(f"Created dataset with ID {dataset.id} with {len(sheets)} Google Sheets")
            return

        for sheet_name, sheet_id in sheets.items():
            print(f"{sheet_name} (ID: {sheet_id})")

    except HttpError as err:
        print(err)


if __name__ == "__main__":
    asyncio.run(main())
