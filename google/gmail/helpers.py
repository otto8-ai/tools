import base64
import os
import gptscript
from filetype import guess_mime
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email import encoders

async def create_message(to, cc, bcc, subject, message_text, attachments):
    if attachments:
        # Create a multipart message when there are attachments
        gptscript_client = gptscript.GPTScript()
        message = MIMEMultipart()
        message.attach(MIMEText(message_text, 'plain'))
    else:
        # Use MIMEText for plain text messages without attachments
        message = MIMEText(message_text, 'plain')

    message['to'] = to
    if cc is not None:
        message['cc'] = cc
    if bcc is not None:
        message['bcc'] = bcc
    message['subject'] = subject

    # Read and attach any workspace files if provided
    for filepath in attachments:
        try:
            # Get the file bytes from the workspace
            wksp_file_path = await prepend_base_path('files', filepath)
            file_content = await gptscript_client.read_file_in_workspace(wksp_file_path)

            # Determine the MIME type and subtype
            mime = guess_mime(file_content) or "application/octet-stream"
            main_type, sub_type = mime.split('/', 1)

            # Create the appropriate MIMEBase object for the attachment
            mime_base = MIMEBase(main_type, sub_type)
            mime_base.set_payload(file_content)
            encoders.encode_base64(mime_base)

            # Add header with the file name
            mime_base.add_header(
                'Content-Disposition',
                f'attachment; filename="{filepath.split("/")[-1]}"'
            )
            message.attach(mime_base)
        except Exception as e:
            # Raise a new exception with the problematic file path included
            raise Exception(f"Error attaching {filepath}: {e}")

    # Encode the message as a base64 string
    raw_message = base64.urlsafe_b64encode(message.as_bytes()).decode('utf-8')
    return {'raw': raw_message}


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


from gptscript.datasets import DatasetElement


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

        try:
            gptscript_client = gptscript.GPTScript()

            elements = []
            for message in all_messages:
                msg_id, msg_str = message_to_string(service, message)
                elements.append(DatasetElement(name=msg_id, description="", contents=msg_str))

            dataset_id = await gptscript_client.add_dataset_elements(
                elements,
                name=f"gmail_{query}",
                description=f"list of emails in Gmail for query {query}"
            )

            print(f"Created dataset with ID {dataset_id} with {len(elements)} emails")
        except Exception as e:
            print("An error occurred while creating the dataset:", e)

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

        try:
            gptscript_client = gptscript.GPTScript()

            elements = []
            for draft in all_drafts:
                draft_id, draft_str = draft_to_string(service, draft)
                elements.append(DatasetElement(name=draft_id, description="", contents=draft_str))

            dataset_id = await gptscript_client.add_dataset_elements(elements, name=f"gmail_drafts")

            print(f"Created dataset with ID {dataset_id} with {len(elements)} drafts")
        except Exception as e:
            print("An error occurred while creating the dataset:", e)

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

async def prepend_base_path(base_path: str, file_path: str):
    """
    Prepend a base path to a file path if it's not already rooted in the base path.

    Args:
        base_path (str): The base path to prepend.
        file_path (str): The file path to check and modify.

    Returns:
        str: The modified file path with the base path prepended if necessary.

    Examples:
      >>> prepend_base_path("files", "my-file.txt")
      'files/my-file.txt'

      >>> prepend_base_path("files", "files/my-file.txt")
      'files/my-file.txt'

      >>> prepend_base_path("files", "foo/my-file.txt")
      'files/foo/my-file.txt'

      >>> prepend_base_path("files", "bar/files/my-file.txt")
      'files/bar/files/my-file.txt'

      >>> prepend_base_path("files", "files/bar/files/my-file.txt")
      'files/bar/files/my-file.txt'
    """
    # Split the file path into parts for checking
    file_parts = os.path.normpath(file_path).split(os.sep)

    # Check if the base path is already at the root
    if file_parts[0] == base_path:
        return file_path

    # Prepend the base path
    return os.path.join(base_path, file_path)
