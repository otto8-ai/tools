import os
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    id = os.getenv("OPPORTUNITY_ID")
    if id is None:
        print("No opportunity id provided")
        exit(1)

    try:
        sf = client()
        sf.Opportunity.delete(id)
        print(f"Opportunity with Id: {id} deleted successfully")
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)


if __name__ == "__main__":
    main()
