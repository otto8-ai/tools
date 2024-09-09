import json
import os

from hubspot import HubSpot
from hubspot.crm.objects import ApiException, SimplePublicObjectInput

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def create_email_engagement(timestamp, owner_id, subject, body, headers):
    engagement = SimplePublicObjectInput(
        properties={
            'hs_timestamp': timestamp,
            'hubspot_owner_id': owner_id,
            'hs_email_direction': "EMAIL",
            'hs_email_subject': subject,
            'hs_email_text': body,
            'hs_email_headers': headers,

        }
    )

    try:
        # Create the email engagement
        engagement_response = client.crm.objects.basic_api.create(object_type='email',
                                                                  simple_public_object_input_for_create=engagement)
        email_id = engagement_response.id
        print(f"Created email engagement with ID: {email_id}")
    except ApiException as e:
        print(f"Error creating email engagement: {e}")


if __name__ == "__main__":
    subject = os.getenv("SUBJECT")
    owner_id = os.getenv("OWNER_ID")
    body = os.getenv("BODY")
    timestamp = int(os.getenv("TIMESTAMP"))

    from_email = os.getenv("FROM_EMAIL")
    from_firstname = os.getenv("FROM_FIRSTNAME")
    from_lastname = os.getenv("FROM_LASTNAME")

    to_firstname = os.getenv("TO_FIRSTNAME")
    to_lastname = os.getenv("TO_LASTNAME")
    to_email = os.getenv("TO_EMAIL")
    to_cc = os.getenv("TO_CC")
    to_bcc = os.getenv("TO_BCC")

    if to_cc is not None:
        to_cc.split(',')
    if to_bcc is not None:
        to_bcc.split(',')

    headers = {
        "from": {
            "email": from_email,
            "firstName": from_firstname,
            "lastName": from_lastname,
        },
        "to": [
            {
                "email": f"{to_firstname} {to_lastname}<{to_email}>",
                "firstName": to_firstname,
                "lastName": to_lastname,
            }
        ],
        "cc": to_cc,
        "bcc": to_bcc,
    }

    create_email_engagement(timestamp, owner_id, subject, body, json.dumps(headers))
