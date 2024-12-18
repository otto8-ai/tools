import json
import os
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    try:
        lead_id = os.getenv("LEAD_ID")
        if lead_id is None:
            print("No lead Id provided")
            exit(1)
        lead = os.getenv("LEAD")
        if lead is None:
            print("No lead provided")
            exit(1)
        lead = json.loads(lead)
    except json.JSONDecodeError:
        print("Invalid JSON provided")
        exit(1)
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)

    try:
        sf = client()
        lead = sf.Lead.update(lead_id, lead)
        print(f"Lead with Id: {lead['id']} updated successfully")
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)


if __name__ == "__main__":
    main()
