import asyncio
import os

import pandas as pd

from auth import gspread_client


async def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    spreadsheet_name = os.getenv('SPREADSHEET_NAME')
    if spreadsheet_id is None and spreadsheet_name is None:
        raise ValueError("Either spreadsheet_id or spreadsheet_name parameter must be set")
    query = os.getenv('QUERY')
    if query is None:
        raise ValueError("query parameter must be set")
    show_columns = os.getenv('SHOW_COLUMNS')
    if show_columns is not None:
        show_columns = show_columns.split(',')
    sheet_name = os.getenv('SHEET_NAME')

    service = gspread_client()
    try:
        spreadsheet = service.open(spreadsheet_name) if spreadsheet_name is not None else service.open_by_key(
            spreadsheet_id)
        if sheet_name is None:
            sheet = spreadsheet.sheet1
        else:
            sheet = spreadsheet.worksheet(sheet_name)
        values = sheet.get_all_records()
        if not values:
            print("No data found.")
            return
        df = pd.DataFrame(values)
        filtered_df = df.query(query)
        # Set the max rows and max columns to display
        pd.set_option('display.max_rows', None)
        if show_columns is None:
            pd.set_option('display.max_columns', 5)
        else:
            pd.set_option('display.max_columns', len(show_columns))
            filtered_df = filtered_df[show_columns]

        print(filtered_df)

    except Exception as err:
        print(err)


def get_cell_reference(row_idx, col_idx):
    col_letter = chr(col_idx + 64)
    return f"{col_letter}{row_idx}"


if __name__ == "__main__":
    asyncio.run(main())
