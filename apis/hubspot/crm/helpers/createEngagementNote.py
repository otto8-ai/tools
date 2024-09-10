import os

from hubspot import HubSpot
from hubspot.crm.objects import ApiException, SimplePublicObjectInput

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def create_note_engagement(timestamp, owner_id, body):
    engagement = SimplePublicObjectInput(
        properties={
            'hs_timestamp': timestamp,
            'hubspot_owner_id': owner_id,
            'hs_note_body': body,
        }
    )

    try:
        # Create the note engagement
        engagement_response = client.crm.objects.basic_api.create(object_type='note',
                                                                  simple_public_object_input_for_create=engagement)
        note_id = engagement_response.id
        print(f"Created note engagement with ID: {note_id}")

    except ApiException as e:
        print(f"Error creating note engagement: {e}")


if __name__ == "__main__":
    owner_id = os.getenv("OWNER_ID")
    body = os.getenv("BODY")
    if len(body) > 65536:
        print("Error: Note body is too long. It must be less than 65536 characters.")
        exit(1)

    timestamp = int(os.getenv("TIMESTAMP"))
    create_note_engagement(timestamp, owner_id, body)
