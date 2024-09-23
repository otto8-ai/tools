import os

from googleapiclient.errors import HttpError

from helpers import client


def main():
    draft_id = os.getenv('DRAFT_ID')
    if draft_id is None:
        raise ValueError("draft_id must be set")

    service = client('gmail', 'v1')
    try:
        sent_message = service.users().drafts().send(userId='me', body={'id': draft_id}).execute()
        print(f"Draft Id: {draft_id} sent successfully! Message Id: {sent_message['id']}")
    except HttpError as err:
        print(err)


if __name__ == "__main__":
    main()
