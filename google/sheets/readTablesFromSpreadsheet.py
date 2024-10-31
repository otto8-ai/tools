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
    if range is None:
        range = "A:Z"

    sheet_name = os.getenv('SHEET_NAME')
    if sheet_name is not None:
        range = f"{sheet_name}!{range}"

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

        tables = []
        current_table = []
        for i, row in enumerate(values):
            if not any(cell.strip() for cell in row):
                if current_table:
                    tables.append(current_table)
                    current_table = []
            else:
                current_table.append(row)
        if current_table:
            tables.append(current_table)

        try:
            gptscript_client = GPTScript()

            dataset = await gptscript_client.create_dataset(
                os.getenv("GPTSCRIPT_WORKSPACE_ID"), f"{spreadsheet.id}_tables", ""
            )

            elements = []
            for index, table in enumerate(tables):
                table_text = "\n".join([f"[{', '.join(row)}]" for row in table])
                elements.append(DatasetElement(name=f"Table {index + 1}", description="", contents=table_text))

            await gptscript_client.add_dataset_elements(os.getenv("GPTSCRIPT_WORKSPACE_ID"), dataset.id, elements)

            print(f"Created dataset with ID {dataset.id} with {len(tables)} tables")
            return
        except Exception as e:
            print("An error occurred while creating the dataset:", e)

    except Exception as err:
        print(err)


if __name__ == "__main__":
    asyncio.run(main())
