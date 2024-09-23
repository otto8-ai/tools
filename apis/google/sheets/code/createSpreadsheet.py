import os

from googleapiclient.errors import HttpError

from auth import client


def main():
    spreadsheet_name = os.getenv('SPREADSHEET_NAME')
    if spreadsheet_name is None:
        raise ValueError("spreadsheet_name is not set")

    props = {
        'properties': {
            'title': spreadsheet_name
        }
    }

    service = client('sheets', 'v4')
    try:
        spreadsheet = service.spreadsheets().create(body=props, fields='spreadsheetId').execute()
        print(f"Spreadsheet named {spreadsheet_name} created with ID: {spreadsheet.get('spreadsheetId')}")
    except HttpError as err:
        print(err)


if __name__ == "__main__":
    main()
