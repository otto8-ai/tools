import os

from helpers import client, list_messages


def main():
    max_results = os.getenv('MAX_RESULTS')
    if max_results is not None:
        max_results = int(max_results)
    query = 'label:inbox'

    service = client('gmail', 'v1')
    list_messages(service, query, max_results)


if __name__ == "__main__":
    main()
