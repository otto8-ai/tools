import base64
import os

from googleapiclient.errors import HttpError

from helpers import client


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

        msg = service.users().messages().get(userId='me', id=email_id, format='full').execute()
        body = get_email_body(msg)
        attachment = has_attachment(msg)

        subject = None
        sender = None
        for header in msg['payload']['headers']:
            if header['name'].lower() == 'subject':
                subject = header['value']
            if header['name'].lower() == 'from':
                sender = header['value']

        print(f'From: {sender}, Subject: {subject}')
        print(f'Body:\n{body}')
        if attachment:
            print('Email has attachment(s)')
            link='https://mail.google.com/mail/u/0/#inbox/' + email_id
            print(f'Link: {link}')

    except HttpError as err:
        print(err)


def has_attachment(message):
    def parse_parts(parts):
        for part in parts:
            if part['filename'] and part['body'].get('attachmentId'):
                return True
        return False

    parts = message['payload'].get('parts', [])
    if parts:
        return parse_parts(parts)
    else:
        return False


def get_email_body(message):
    def parse_parts(parts):
        for part in parts:
            mime_type = part['mimeType']
            if mime_type == 'text/plain' or mime_type == 'text/html':
                body_data = part['body']['data']
                decoded_body = base64.urlsafe_b64decode(body_data).decode('utf-8')
                return decoded_body
            if mime_type == 'multipart/alternative' or mime_type == 'multipart/mixed':
                return parse_parts(part['parts'])
        return None

    try:
        parts = message['payload'].get('parts', [])
        if parts:
            return parse_parts(parts)
        else:
            body_data = message['payload']['body']['data']
            decoded_body = base64.urlsafe_b64decode(body_data).decode('utf-8')
            return decoded_body
    except Exception as e:
        print(f'Error while decoding the email body: {e}')
        return None


if __name__ == "__main__":
    main()
