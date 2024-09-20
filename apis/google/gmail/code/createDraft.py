import os

from googleapiclient.errors import HttpError

from helpers import client, create_message


def main():
    to_emails = os.getenv('TO_EMAILS')
    if to_emails is None:
        raise ValueError("At least one recipient must be specified with 'to_emails'")

    cc_emails = os.getenv('CC_EMAILS')
    bcc_emails = os.getenv('BCC_EMAILS')
    subject = os.getenv('SUBJECT')
    if subject is None:
        raise ValueError("Email subject must be set")

    message = os.getenv('MESSAGE')
    if message is None:
        raise ValueError("Email message must be set")

    service = client('gmail', 'v1')
    try:
        create_draft(service=service, to=to_emails, cc=cc_emails, bcc=bcc_emails, subject=subject, body=message)
    except HttpError as err:
        print(err)


def create_draft(service, to, cc, bcc, subject, body):
    try:
        message = create_message(to=to, cc=cc, bcc=bcc, subject=subject, message_text=body)

        draft = {
            'message': message
        }

        draft_response = service.users().drafts().create(userId='me', body=draft).execute()
        print(f"Draft Id: {draft_response['id']} - Draft created successfully!")
    except HttpError as error:
        print(f'An error occurred: {error}')


if __name__ == "__main__":
    main()
