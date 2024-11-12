import json
import os
import subprocess
import sys
from dataclasses import dataclass

from azure.identity import DefaultAzureCredential
from azure.mgmt.cognitiveservices import CognitiveServicesManagementClient
from azure.mgmt.resource import ResourceManagementClient
from openai import AzureOpenAI

endpoint: str
api_key: str
deployment_name: str


async def list_resource_groups(client: ResourceManagementClient):
    group_list = client.resource_groups.list()

    column_width = 40
    print("Resource Group".ljust(column_width) + "Location", file=sys.stderr)
    print("-" * (column_width * 2), file=sys.stderr)
    for group in list(group_list):
        print(f"{group.name:<{column_width}}{group.location}", file=sys.stderr)
    print("", file=sys.stderr)


async def list_openai(client: CognitiveServicesManagementClient, resource_group: str):
    accounts = client.accounts.list_by_resource_group(resource_group_name=resource_group, api_version="2023-05-01")

    column_width = 40
    print(f"OpenAI Endpoints in {resource_group}".ljust(column_width) + "Model Name", file=sys.stderr)
    print("-" * (column_width * 2), file=sys.stderr)
    for account in list(accounts):
        if account.kind == "OpenAI":
            deployments = client.deployments.list(resource_group_name=resource_group,
                                                  account_name=account.name, api_version="2023-05-01")
            deployments = list(deployments)
            model_id = deployments[0].properties.model.name
            print(f"{account.name:<{column_width}}{model_id}", file=sys.stderr)
    print("", file=sys.stderr)


async def get_api_key(resource, resource_group: str,
                      client: CognitiveServicesManagementClient) -> str:
    keys = client.accounts.list_keys(resource_group, resource.name)
    return keys.key1


@dataclass
class AzureConfig:
    endpoint: str
    deployment_name: str
    api_key: str

    def to_json(self):
        return json.dumps(
            self,
            default=lambda o: o.__dict__,
            sort_keys=True)


async def get_azure_config(model_name: str | None = None,
                           subscription_id: str | None = None,
                           resource_group: str | None = None) -> AzureConfig | None:
    global endpoint
    global api_key
    global deployment_name

    if 'OTTO8_AZURE_OPENAI_MODEL_PROVIDER_ENDPOINT' in os.environ and 'OTTO8_AZURE_OPENAI_MODEL_PROVIDER_API_KEY' in os.environ and 'OTTO8_AZURE_OPENAI_MODEL_PROVIDER_DEPLOYMENT_NAME' in os.environ:
        endpoint = os.environ['OTTO8_AZURE_OPENAI_MODEL_PROVIDER_ENDPOINT']
        api_key = os.environ['OTTO8_AZURE_OPENAI_MODEL_PROVIDER_API_KEY']
        deployment_name = os.environ['OTTO8_AZURE_OPENAI_MODEL_PROVIDER_DEPLOYMENT_NAME']

    if 'endpoint' in globals() and 'api_key' in globals() and 'deployment_name' in globals():
        return AzureConfig(
            endpoint=endpoint,
            api_key=api_key,
            deployment_name=deployment_name
        )

    try:
        credential = DefaultAzureCredential()
    except:
        print("Failed to get Azure credentials.", file=sys.stderr)
        sys.exit(1)

    if subscription_id is None:
        return None

    resource_client = ResourceManagementClient(credential=credential, subscription_id=subscription_id)
    cognitive_client = CognitiveServicesManagementClient(credential=credential, subscription_id=subscription_id)
    model_id: str

    if resource_group is None:
        await list_resource_groups(resource_client)
        print("Please select an Azure Resource Group.", file=sys.stderr)
        return None

    accounts = cognitive_client.accounts.list_by_resource_group(resource_group_name=resource_group,
                                                                api_version="2023-05-01")
    for account in list(accounts):
        selected_resource = account
        endpoint = account.properties.endpoint
        deployments = cognitive_client.deployments.list(resource_group_name=resource_group,
                                                        account_name=account.name, api_version="2023-05-01")
        for deployment in list(deployments):
            if deployment.properties.model.name == model_name:
                deployment_name = deployment.name
                model_id = deployment.properties.model.name
                break

    if 'model_id' not in locals():
        print(f"Did not find any matches for model name {model_name}.", file=sys.stderr)
        sys.exit(1)

    api_key = await get_api_key(client=cognitive_client, resource=selected_resource, resource_group=resource_group)

    return AzureConfig(endpoint,
                       deployment_name,
                       api_key,
                       )


def client(endpoint: str, deployment_name: str, api_key: str, api_version: str = "2024-02-01") -> AzureOpenAI:
    return AzureOpenAI(
        azure_endpoint=endpoint,
        azure_deployment=deployment_name,
        api_key=api_key,
        api_version=api_version
    )
