import os

from hubspot import HubSpot

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)

try:
    user = client.auth.oauth.access_tokens_api.get(token)
    print(f"Current user is {user.user}. UserId is {user.user_id}.")
except Exception as e:
    print(f"Error getting current user information: {e}")


owners = client.crm.owners.owners_api.get_page(email=user.user)
print(f"OwnerId is {owners.results[0].id}. Use this to look up ownership associations.")
