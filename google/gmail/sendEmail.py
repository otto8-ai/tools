import os
import asyncio
from googleapiclient.errors import HttpError

from helpers import client, create_message


async def main():
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

    attachments = os.getenv('ATTACHMENTS', '').split(',')
    attachments = [attachment.strip() for attachment in attachments if attachment.strip()]

    service = client('gmail', 'v1')
    try:
        await send_message(
            service=service,
            to=to_emails,
            cc=cc_emails,
            bcc=bcc_emails,
            subject=subject,
            body=message,
            attachments=attachments
        )
    except HttpError as err:
        print(err)
    except Exception as err:
        print(err)


async def send_message(service, to, cc, bcc, subject, body, attachments):
    try:
        message = await create_message(
            to=to,
            cc=cc,
            bcc=bcc,
            subject=subject,
            message_text=body,
            attachments=attachments
        )
        sent_message = service.users().messages().send(userId='me', body=message).execute()
        print(f"Message Id: {sent_message['id']} - Message sent successfully!")
    except HttpError as error:
        print(f'An error occurred: {error}')


if __name__ == "__main__":
    asyncio.run(main())
