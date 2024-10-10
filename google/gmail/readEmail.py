import base64
import os

from googleapiclient.errors import HttpError

from helpers import client, fetch_email_or_draft, extract_message_headers, get_email_body, has_attachment


def main():
    email_id = os.getenv('EMAIL_ID')
    email_subject = os.getenv('EMAIL_SUBJECT')
    if email_id is None and email_subject is None:
        raise ValueError("Either email_id or email_subject must be set")

    service = client('gmail', 'v1')
    try:
        if email_subject is not None:
            query = f'subject:"{email_subject}"'
            response = service.users().messages().list(userId='me', q=query).execute()
            if not response:
                raise ValueError(f"No emails found with subject: {email_subject}")
            email_id = response['messages'][0]['id']

        msg = fetch_email_or_draft(service, email_id)
        body = get_email_body(msg)
        attachment = has_attachment(msg)

        subject, sender, to, cc, bcc, date = extract_message_headers(msg)

        print(f'From: {sender}, Subject: {subject}, To: {to}, CC: {cc}, Bcc: {bcc}, Date: {date}')
        print(f'Body:\n{body}')
        if attachment:
            print('Email has attachment(s)')
            link='https://mail.google.com/mail/u/0/#inbox/' + email_id
            print(f'Link: {link}')

    except HttpError as err:
        print(err)



if __name__ == "__main__":
    main()
