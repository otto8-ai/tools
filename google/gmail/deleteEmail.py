import os

from googleapiclient.errors import HttpError

from helpers import client


def main():
    email_id = os.getenv('EMAIL_ID')
    if email_id is None:
        raise ValueError("email_id must be set")

    service = client('gmail', 'v1')
    try:
        service.users().messages().trash(userId='me', id=email_id).execute()
        print(f"Email Id: {email_id} deleted successfully!")
    except HttpError as err:
        print(err)


if __name__ == "__main__":
    main()
