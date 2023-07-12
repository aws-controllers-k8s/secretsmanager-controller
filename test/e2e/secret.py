import boto3

client = boto3.client('secretsmanager')


def get_tags(secret_name):
    response = client.describe_secret(SecretId=secret_name)
    return response['Tags']


def get_deleted_date(secret_name):
    response = client.describe_secret(SecretId=secret_name)
    return response['DeletedDate']


def get_secret_value(secret_name):
    response = client.get_secret_value(SecretId=secret_name)
    return response['SecretString']
