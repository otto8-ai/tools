import os

from auth import gspread_client


def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    spreadsheet_name = os.getenv('SPREADSHEET_NAME')
    cell = os.getenv('CELL')
    if cell is None:
        raise ValueError("cell parameter must be set")
    if spreadsheet_id is None and spreadsheet_name is None:
        raise ValueError("Either spreadsheet_id or spreadsheet_name parameter must be set")

    data = os.getenv('DATA')
    if data is None:
        raise ValueError("data parameter must be set")

    service = gspread_client()
    try:
        spreadsheet = service.open(spreadsheet_name) if spreadsheet_name is not None else service.open_by_key(
            spreadsheet_id)
        sheet = spreadsheet.sheet1
        sheet.update_acell(cell, data)
    except Exception as err:
        print(err)

    print("Data written successfully")


if __name__ == "__main__":
    main()
