import json
import os
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    try:
        contact = os.getenv("CONTACT")
        if contact is None:
            print("No contact provided")
            exit(1)
        contact = json.loads(contact)
    except json.JSONDecodeError:
        print("Invalid JSON provided")
        exit(1)
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)

    try:
        sf = client()
        contact = sf.Contact.create(contact)
        print(f"Contact created successfully with Id: {contact['id']}")
    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)


if __name__ == "__main__":
    main()
