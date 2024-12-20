import json
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    sf = client()
    desc = sf.Lead.describe()
    print("Available fields for Lead objects:")
    for field in desc["fields"]:
        if field["name"] not in [
            "Id",
            "IsDeleted",
            "MasterRecordId",
            "SystemModstamp",
            "LastViewedDate",
            "LastReferencedDate",
            "Jigsaw",
            "JigsawContactId",
            "CleanStatus",
            "EmailBouncedReason",
            "EmailBouncedDate",
            "OwnerId",
            "SICCode__c",
            "Latitude",
            "Longitude",
            "GeoCodeAccuracy",
            "CurrentGenerators__c",
        ]:
            print(
                f"name: {field['name']} - label: {field['label']} - type: {field['type']} - picklistValues: {json.dumps(field.get('picklistValues'))}"
            )


if __name__ == "__main__":
    main()
