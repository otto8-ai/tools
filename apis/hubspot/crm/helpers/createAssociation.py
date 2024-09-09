import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException

from getAssociationTypeIdFromObjectToObject import get_association_type_id_from_object_to_object

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def create_association(from_object_type, from_object_id, to_object_type, to_object_id):
    resp = get_association_type_id_from_object_to_object(from_object_type, to_object_type)

    association_spec = [{
        "associationCategory": resp.results[0].category,
        "associationTypeId": resp.results[0].type_id,
    }]

    try:
        client.crm.associations.v4.basic_api.create(object_type=from_object_type, object_id=from_object_id,
                                                    to_object_type=to_object_type, to_object_id=to_object_id,
                                                    association_spec=association_spec)
    except ApiException as e:
        print("Exception when calling associations->create: %s\n" % e)
        print("Try mapping the objects the other way around.")

        resp = get_association_type_id_from_object_to_object(to_object_type, from_object_type)
        association_spec = [{
            "associationCategory": resp.results[0].category,
            "associationTypeId": resp.results[0].type_id,
        }]
        try:
            client.crm.associations.v4.basic_api.create(object_type=to_object_type, object_id=to_object_id,
                                                        to_object_type=from_object_type, to_object_id=from_object_id,
                                                        association_spec=association_spec)
        except ApiException as e:
            print("Exception when calling associations->create: %s\n" % e)
            print("The reverse mapping didn't work either, there is something else wrong.")
            exit(1)


if __name__ == "__main__":
    from_object_type = os.getenv("FROM_OBJECT_TYPE")
    from_object_id = os.getenv("FROM_OBJECT_ID")

    to_object_type = os.getenv("TO_OBJECT_TYPE")
    to_object_id = os.getenv("TO_OBJECT_ID")
    create_association(from_object_type, from_object_id, to_object_type, to_object_id)
