from googleapiclient.errors import HttpError

from auth import client


def main():
    service = client('drive', 'v3')
    try:
        # Query Google Drive for all Google Sheets files
        query = "mimeType='application/vnd.google-apps.spreadsheet'"
        page_token = None
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
                print(f"{item['name']} (ID: {item['id']})")

            page_token = results.get('nextPageToken')
            if not page_token:
                break

    except HttpError as err:
        print(err)


if __name__ == "__main__":
    main()
