import json
import os
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    try:
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
        lead = sf.Lead.create(lead)
        print(f"Lead created successfully with Id: {lead['id']}")
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)


if __name__ == "__main__":
    main()
