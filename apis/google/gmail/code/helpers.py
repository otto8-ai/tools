import base64
from email.mime.text import MIMEText


def create_message(to, cc, bcc, subject, message_text):
    message = MIMEText(message_text)
    message['to'] = to
    if cc is not None:
        message['cc'] = cc
    if bcc is not None:
        message['bcc'] = bcc
    message['subject'] = subject
    raw_message = base64.urlsafe_b64encode(message.as_bytes()).decode('utf-8')
    return {'raw': raw_message}


import os

from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError


def client(service_name: str, version: str):
    token = os.getenv('GMAIL_GOOGLE_OAUTH_TOKEN')
    if token is None:
        raise ValueError("GMAIL_GOOGLE_OAUTH_TOKEN environment variable is not set")

    creds = Credentials(token=token)
    try:
        service = build(serviceName=service_name, version=version, credentials=creds)
        return service
    except HttpError as err:
        print(err)
        exit(1)


def list_messages(service, query, max_results):
    all_messages = []
    next_page_token = None
    try:
        while True:
            if next_page_token:
                results = service.users().messages().list(userId='me', q=query, pageToken=next_page_token,
                                                          maxResults=10).execute()
            else:
                results = service.users().messages().list(userId='me', q=query, maxResults=10).execute()
            messages = results.get('messages', [])
            if not messages:
                break

            all_messages.extend(messages)
            if max_results is not None and len(all_messages) >= max_results:
                break

            next_page_token = results.get('nextPageToken')
            if not next_page_token:
                break

        display_list_messages(service, all_messages)

    except HttpError as err:
        print(err)


from datetime import datetime, timezone


def display_list_messages(service, messages: list):
    print('Messages:')
    for message in messages:
        msg = (service.users().messages().get(userId='me',
                                              id=message['id'],
                                              format='metadata',
                                              metadataHeaders=['From', 'Subject'])
               .execute())

        msg_id = msg['id']
        subject = None
        sender = None
        date = datetime.fromtimestamp(int(msg['internalDate']) / 1000, timezone.utc).astimezone().strftime(
            '%Y-%m-%d %H:%M:%S')

        for header in msg['payload']['headers']:
            if header['name'] == 'Subject':
                subject = header['value']
            if header['name'] == 'From':
                sender = header['value']

        print(f"ID: {msg_id} From: {sender}, Subject: {subject}, Received: {date}")
