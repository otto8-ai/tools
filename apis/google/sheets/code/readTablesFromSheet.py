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
        # Call the Sheets API
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

        # Variables to store the start and end of each table
        tables = []
        current_table = []

        # Iterate over each row in the sheet
        for i, row in enumerate(values):
            # Check if the row is completely empty
            if not any(cell.strip() for cell in row):
                # If there's a current table being built, finalize it
                if current_table:
                    tables.append(current_table)
                    current_table = []
            else:
                # Non-empty row, so add it to the current table
                current_table.append(row)

        # If there's a table still being built at the end, add it
        if current_table:
            tables.append(current_table)

        # Output the detected tables
        for index, table in enumerate(tables):
            print(f"Table {index + 1}:")
            for row in table:
                print(row)
            print("-" * 40)
    except HttpError as err:
        print(err)


if __name__ == "__main__":
    main()
