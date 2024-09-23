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

    draft_id = os.getenv('DRAFT_ID')
    if draft_id is None:
        raise ValueError("draft_id must be set")

    service = client('gmail', 'v1')
    try:
        update_draft(service=service, draft_id=draft_id, to=to_emails, cc=cc_emails, bcc=bcc_emails, subject=subject,
                     body=message)
    except HttpError as err:
        print(err)


def update_draft(service, draft_id, to, cc, bcc, subject, body):
    try:
        message = create_message(to=to, cc=cc, bcc=bcc, subject=subject, message_text=body)
        updated_draft = {
            'id': draft_id,
            'message': message,
        }

        draft_response = service.users().drafts().update(userId='me', id=draft_id, body=updated_draft).execute()
        print(f"Draft Id: {draft_response['id']} - Draft updated successfully!")
    except HttpError as error:
        print(f'An error occurred: {error}')


if __name__ == "__main__":
    main()
