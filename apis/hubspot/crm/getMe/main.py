import os

from hubspot import HubSpot

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot()

try:
    user = client.auth.oauth.access_tokens_api.get(token)
    print(f"Current user is {user.user}. User Id is {user.user_id}.")
except Exception as e:
    print(f"Error getting current user information: {e}")
