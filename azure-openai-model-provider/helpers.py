from azure.mgmt.cognitiveservices import CognitiveServicesManagementClient

def list_openai(client: CognitiveServicesManagementClient, resource_group: str):
    accounts = client.accounts.list_by_resource_group(resource_group_name=resource_group, api_version="2023-05-01")
    deployments = []
    for account in accounts:
        if account.kind == "OpenAI":
            deployments.extend(client.deployments.list(
                resource_group_name=resource_group,
                account_name=account.name,
                api_version="2023-05-01",
            ))

    return deployments