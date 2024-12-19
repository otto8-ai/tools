import json
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    sf = client()
    desc = sf.Contact.describe()
    print("Available fields for Contact objects:")
    for field in desc["fields"]:
        if field["name"] not in [
            "OwnerId",
            "IsDeleted",
            "MasterRecordId",
            "AccountId",
            "IsEmailBounced",
            "EmailBouncedReason",
            "EmailBouncedDate",
            "EmailBounced",
            "Jigsaw",
            "JigsawContactId",
            "CleanStatus",
            "IndividualId",
            "Level__c",
            "Languages__c",
            "CreatedDate",
            "CreatedById",
            "LastModifiedDate",
            "LastModifiedById",
            "SystemModstamp",
            "LastActivityDate",
            "LastCURequestDate",
            "LastCUUpdateDate",
            "LastViewedDate",
            "LastReferencedDate",
            "ReportsToId",
        ]:
            print(
                f"name: {field['name']} - label: {field['label']} - type: {field['type']} - picklistValues: {json.dumps(field.get('picklistValues'))}"
            )


if __name__ == "__main__":
    main()
