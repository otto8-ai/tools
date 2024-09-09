import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_association_type_id_from_object_to_object(from_object_type, to_object_type):
    try:
        resp = client.crm.associations.v4.schema.definitions_api.get_all(from_object_type=from_object_type,
                                                                         to_object_type=to_object_type)
    except ApiException as e:
        print("Exception when calling associations->get_all: %s\n" % e)
        print("This should not happen, there is something wrong with the request.")
        exit(1)

    print(
        f"The association type ID from {from_object_type} to {to_object_type} is {resp.results[0].type_id}. This is a {resp.results[0].category} association.")
    return resp


if __name__ == "__main__":
    from_object_type = os.getenv("FROM_OBJECT_TYPE")
    to_object_type = os.getenv("TO_OBJECT_TYPE")
    get_association_type_id_from_object_to_object(from_object_type, to_object_type)
