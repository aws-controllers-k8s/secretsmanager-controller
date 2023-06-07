import boto3


def get_tags(secret_name):
    client = boto3.client('secretsmanager')
    response = client.describe_secret(SecretId=secret_name)
    return response['Tags']


def get_deleted_date(secret_name):
    client = boto3.client('secretsmanager')
    response = client.describe_secret(SecretId=secret_name)
    return response['DeletedDate']
