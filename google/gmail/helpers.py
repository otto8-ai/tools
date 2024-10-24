import base64
import sys
from email.mime.text import MIMEText

import gptscript


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
    token = os.getenv('GOOGLE_OAUTH_TOKEN')
    if token is None:
        raise ValueError("GOOGLE_OAUTH_TOKEN environment variable is not set")

    creds = Credentials(token=token)
    try:
        service = build(serviceName=service_name, version=version, credentials=creds)
        return service
    except HttpError as err:
        print(err)
        exit(1)


async def list_messages(service, query, max_results):
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

        if len(all_messages) > 10:
            gptscript_client = gptscript.GPTScript()
            try:
                dataset = await gptscript_client.create_dataset(
                    os.getenv("GPTSCRIPT_WORKSPACE_DIR"),
                    f"gmail_{query}",
                    f"list of emails in Gmail for query {query}")
                for message in all_messages:
                    msg_id, msg_str = message_to_string(service, message)
                    await gptscript_client.add_dataset_element(
                        os.getenv("GPTSCRIPT_WORKSPACE_DIR"),
                        dataset.id,
                        msg_id,
                        msg_str
                    )
                print(f"Created dataset with ID {dataset.id} with {len(all_messages)} emails")
                return
            except Exception as e:
                print("An error occurred while creating the dataset:", e, file=sys.stderr)
                pass  # Ignore errors if we got any, and just print the results.

        display_list_messages(service, all_messages)

    except HttpError as err:
        print(err)


from datetime import datetime, timezone


def message_to_string(service, message):
    msg = (service.users().messages().get(userId='me',
                                          id=message['id'],
                                          format='metadata',
                                          metadataHeaders=['From', 'Subject'])
           .execute())
    msg_id = msg['id']
    subject, sender, to, cc, bcc, date = extract_message_headers(msg)
    return msg_id, f"ID: {msg_id} From: {sender}, Subject: {subject}, To: {to}, CC: {cc}, Bcc: {bcc}, Received: {date}"


def display_list_messages(service, messages: list):
    print('Messages:')
    for message in messages:
        _, msg_str = message_to_string(service, message)
        print(msg_str)


async def list_drafts(service, max_results=None):
    all_drafts = []
    next_page_token = None
    try:
        while True:
            if next_page_token:
                results = service.users().drafts().list(userId='me', pageToken=next_page_token, maxResults=10).execute()
            else:
                results = service.users().drafts().list(userId='me', maxResults=10).execute()

            drafts = results.get('drafts', [])
            if not drafts:
                break

            all_drafts.extend(drafts)
            if max_results is not None and len(all_drafts) >= max_results:
                break

            next_page_token = results.get('nextPageToken')
            if not next_page_token:
                break

        if len(all_drafts) > 10:
            try:
                gptscript_client = gptscript.GPTScript()
                dataset = await gptscript_client.create_dataset(
                    os.getenv("GPTSCRIPT_WORKSPACE_DIR"),
                    "gmail_drafts",
                    "list of drafts in Gmail")
                for draft in all_drafts:
                    draft_id, draft_str = draft_to_string(service, draft)
                    await gptscript_client.add_dataset_element(
                        os.getenv("GPTSCRIPT_WORKSPACE_DIR"),
                        dataset.id,
                        draft_id,
                        draft_str
                    )
                print(f"Created dataset with ID {dataset.id} with {len(all_drafts)} drafts")
                return
            except Exception as e:
                print("An error occurred while creating the dataset:", e, file=sys.stderr)
                pass  # Ignore errors if we got any, and just print the results.

        display_list_drafts(service, all_drafts)

    except HttpError as err:
        print(f"An error occurred: {err}")


def draft_to_string(service, draft):
    draft_id = draft['id']
    draft_msg = service.users().drafts().get(userId='me', id=draft_id).execute()
    msg = draft_msg['message']
    subject, sender, to, cc, bcc, date = extract_message_headers(msg)
    return draft_id, f"Draft ID: {draft_id}, From: {sender}, Subject: {subject}, To: {to}, CC: {cc}, Bcc: {bcc}, Saved: {date}"


def display_list_drafts(service, drafts: list):
    print('Drafts:')
    for draft in drafts:
        _, draft_str = draft_to_string(service, draft)
        print(draft_str)


def extract_message_headers(message):
    subject = None
    sender = None
    to = None
    cc = None
    bcc = None
    date = None

    if message is not None:
        for header in message['payload']['headers']:
            if header['name'].lower() == 'subject':
                subject = header['value']
            if header['name'].lower() == 'from':
                sender = header['value']
            if header['name'].lower() == 'to':
                to = header['value']
            if header['name'].lower() == 'cc':
                cc = header['value']
            if header['name'].lower() == 'bcc':
                bcc = header['value']
            date = datetime.fromtimestamp(int(message['internalDate']) / 1000, timezone.utc).astimezone().strftime(
                '%Y-%m-%d %H:%M:%S')

    return subject, sender, to, cc, bcc, date


def fetch_email_or_draft(service, obj_id):
    try:
        # Try fetching as an email first
        return service.users().messages().get(userId='me', id=obj_id, format='full').execute()
    except HttpError as email_err:
        if email_err.resp.status == 404 or email_err.resp.status == 400:
            # If email not found, try fetching as a draft
            draft_msg = service.users().drafts().get(userId='me', id=obj_id).execute()
            return draft_msg['message']
        else:
            raise email_err  # Reraise the error if it's not a 404 (not found)


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
