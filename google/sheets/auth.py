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


def gspread_client():
    import gspread
    token = os.getenv('GOOGLE_OAUTH_TOKEN')
    if token is None:
        raise ValueError("GOOGLE_OAUTH_TOKEN environment variable is not set")

    creds = Credentials(token=token)
    try:
        service = gspread.authorize(creds)

        return service
    except HttpError as err:
        print(err)
        exit(1)
