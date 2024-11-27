import asyncio
import base64
import gzip
import json
import os
from io import StringIO

import pandas as pd


async def main():
    data = os.getenv('DATA')
    try:
        data = json.loads(data)
        decoded = base64.b64decode(data['_gz'])
        decompressed = gzip.decompress(decoded)
        values = StringIO(decompressed.decode('utf-8'))
    except:
        values = StringIO(data)

    query = os.getenv('QUERY')
    if query is None:
        raise ValueError("query parameter must be set")
    show_columns = os.getenv('SHOW_COLUMNS')
    if show_columns == '':
        show_columns = None
    if show_columns is not None:
        show_columns = [item.strip() for item in show_columns.split(',')]
    else:
        show_columns = None

    try:
        df = pd.read_csv(values)
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


if __name__ == "__main__":
    asyncio.run(main())
