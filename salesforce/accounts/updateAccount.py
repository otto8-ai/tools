import json
import os
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    try:
        account_id = os.getenv("ACCOUNT_ID")
        if account_id is None:
            print("No account Id provided")
            exit(1)
        account = os.getenv("ACCOUNT")
        if account is None:
            print("No account provided")
            exit(1)
        account = json.loads(account)
    except json.JSONDecodeError:
        print("Invalid JSON provided")
        exit(1)
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)

    try:
        sf = client()
        account = sf.Account.update(account_id, account)
        print(f"Account with Id: {account['id']} updated successfully")
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)


if __name__ == "__main__":
    main()
