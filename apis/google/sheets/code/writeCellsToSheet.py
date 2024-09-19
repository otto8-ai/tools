import csv
import io
import os

from gspread.exceptions import APIError

from auth import gspread_client


def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    spreadsheet_name = os.getenv('SPREADSHEET_NAME')
    if spreadsheet_id is None and spreadsheet_name is None:
        raise ValueError("Either spreadsheet_id or spreadsheet_name must be set")

    raw_data = os.getenv('DATA')
    if raw_data is None:
        raise ValueError("data must be set")
    else:
        data_csv_io = io.StringIO(raw_data)
        data_csv = csv.reader(data_csv_io)
        data = [row for row in data_csv]


    service = gspread_client()
    try:
        spreadsheet = service.open(spreadsheet_name) if spreadsheet_name is not None else service.open_by_key(
            spreadsheet_id)
        sheet = spreadsheet.sheet1
        sheet.append_rows(data)
    except APIError as err:
        print(err)

    print("Data written successfully")


if __name__ == "__main__":
    main()
