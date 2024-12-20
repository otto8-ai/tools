import json
import os


from helpers.auth import client


def main():
    query = os.getenv("QUERY")
    if query is None:
        print("No query provided")
        exit(1)

    sf = client()
    res = sf.query(query)
    if res.get("totalSize") > 0:
        print_res(res)
        next_url = res.get("nextRecordsUrl")
        while next_url:
            res = sf.query_more(next_url, True)
            next_url = res.get("nextRecordsUrl")
            print_res(res)
    else:
        print("No records found")
        exit(0)


def print_res(res):
    for record in res["records"]:
        print(json.dumps(record))


if __name__ == "__main__":
    main()
