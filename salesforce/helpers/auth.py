import os

from simple_salesforce import Salesforce


def client() -> Salesforce:
    access_token = os.environ.get("GPTSCRIPT_SALESFORCE_TOKEN")
    instance_url = os.environ.get("GPTSCRIPT_SALESFORCE_URL")
    return Salesforce(instance_url=instance_url, session_id=access_token)
