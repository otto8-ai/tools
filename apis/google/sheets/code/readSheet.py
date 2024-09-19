import os

from googleapiclient.errors import HttpError

from auth import client


def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    if spreadsheet_id is None:
        raise ValueError("spreadsheet_id is not set")
    range = os.getenv('RANGE')
    if range is None:
        range = "A:Z"
    sheet_name = os.getenv('SHEET_NAME')
    if sheet_name is not None:
        range = f"{sheet_name}!{range}"

    service = client('sheets', 'v4')
    try:
        sheet = service.spreadsheets()
        result = (
            sheet.values()
            .get(spreadsheetId=spreadsheet_id, range=range)
            .execute()
        )
        values = result.get("values", [])

        if not values:
            print("No data found.")
            return

        for row in values:
            print(', '.join(map(str, row)))
    except HttpError as err:
        print(err)


if __name__ == "__main__":
    main()
