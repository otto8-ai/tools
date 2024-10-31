import asyncio
import os

from gptscript import GPTScript
from gptscript.datasets import DatasetElement

from auth import gspread_client


async def main():
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

        cell_values = dict()
        for row_idx, row in enumerate(values, start=1):
            for col_idx, value in enumerate(row, start=1):
                cell_reference = get_cell_reference(row_idx, col_idx)
                cell_values[cell_reference] = value

        try:
            gptscript_client = GPTScript()

            dataset = await gptscript_client.create_dataset(
                os.getenv("GPTSCRIPT_WORKSPACE_ID"),
                f"{spreadsheet.id}_data",
                f"data for Google Sheet with ID {spreadsheet.id}",
            )

            elements = []
            for cell_reference, value in cell_values.items():
                elements.append(DatasetElement(
                    name=cell_reference,
                    description="",
                    contents=value if value != "" else " "
                ))

            await gptscript_client.add_dataset_elements(os.getenv("GPTSCRIPT_WORKSPACE_ID"), dataset.id, elements)

            print(f"Dataset created with ID {dataset.id} with {len(cell_values)} elements")
        except Exception as e:
            print("An error occurred while creating the dataset:", e)

    except Exception as err:
        print(err)


def get_cell_reference(row_idx, col_idx):
    col_letter = chr(col_idx + 64)
    return f"{col_letter}{row_idx}"


if __name__ == "__main__":
    asyncio.run(main())
