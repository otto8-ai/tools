import os

from googleapiclient.errors import HttpError

from auth import client


def main():
    sheet_name = os.getenv('SHEET_NAME')
    if sheet_name is None:
        raise ValueError("sheet_name is not set")

    props = {
        'properties': {
            'title': sheet_name
        }
    }

    service = client('sheets', 'v4')
    try:
        spreadsheet = service.spreadsheets().create(body=props, fields='spreadsheetId').execute()
        print(f"Spreadsheet named {sheet_name} created with ID: {spreadsheet.get('spreadsheetId')}")
    except HttpError as err:
        print(err)


if __name__ == "__main__":
    main()
