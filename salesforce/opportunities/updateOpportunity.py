import json
import os
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    try:
        opportunity_id = os.getenv("OPPORTUNITY_ID")
        if opportunity_id is None:
            print("No opportunity Id provided")
            exit(1)
        opportunity = os.getenv("OPPORTUNITY")
        if opportunity is None:
            print("No opportunity provided")
            exit(1)
        opportunity = json.loads(opportunity)
    except json.JSONDecodeError:
        print("Invalid JSON provided")
        exit(1)
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)

    try:
        sf = client()
        opportunity = sf.Opportunity.update(opportunity_id, opportunity)
        print(f"Opportunity with Id: {opportunity['id']} updated successfully")
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)


if __name__ == "__main__":
    main()
