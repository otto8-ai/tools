import json
import os
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    try:
        case = os.getenv("CASE")
        if case is None:
            print("No case provided")
            exit(1)
        case = json.loads(case)
    except json.JSONDecodeError:
        print("Invalid JSON provided")
        exit(1)
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)

    try:
        sf = client()
        case = sf.Case.create(case)
        print(f"Case created successfully with Id: {case['id']}")
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)


if __name__ == "__main__":
    main()
