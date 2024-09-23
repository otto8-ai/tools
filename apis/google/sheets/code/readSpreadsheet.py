import os

from auth import gspread_client


def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    spreadsheet_name = os.getenv('SPREADSHEET_NAME')
    if spreadsheet_id is None and spreadsheet_name is None:
        raise ValueError("Either spreadsheet_id or spreadsheet_name parameter must be set")
    range = os.getenv('RANGE')
    sheet_name = os.getenv('SHEET_NAME')
    show_cell_ref = os.getenv('SHOW_CELL_REF', 'true').lower() == 'true'

    service = gspread_client()
    try:
        spreadsheet = service.open(spreadsheet_name) if spreadsheet_name is not None else service.open_by_key(
            spreadsheet_id)
        if sheet_name is None:
            sheet = spreadsheet.sheet1
        else:
            sheet = spreadsheet.worksheet(sheet_name)
        if range is None:
            values = sheet.get_all_values()
        else:
            values = sheet.get(range)

        if not values:
            print("No data found.")
            return
        if show_cell_ref:
            for row_idx, row in enumerate(values, start=1):
                for col_idx, value in enumerate(row, start=1):
                    cell_reference = get_cell_reference(row_idx, col_idx)
                    print(f"{cell_reference}: {value}")
        else:
            for row in values:
                print(row)
    except Exception as err:
        print(err)


def get_cell_reference(row_idx, col_idx):
    col_letter = chr(col_idx + 64)
    return f"{col_letter}{row_idx}"


if __name__ == "__main__":
    main()
