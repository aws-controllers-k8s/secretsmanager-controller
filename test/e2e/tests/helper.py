from acktest import tags


class SecretsManagerValidator:
    def __init__(self, secretsmanager_client):
        self.secretsmanager_client = secretsmanager_client

    def assert_tags(self, secret_name, expected_tags):
        response = self.secretsmanager_client.describe_secret(SecretId=secret_name)
        actual_tags = response['Tags']

        tags.assert_equal_without_ack_tags(
            expected=expected_tags,
            actual=actual_tags,
        )

    def assert_secret_value(self, secret_name, expected_value):
        response = self.secretsmanager_client.get_secret_value(SecretId=secret_name)
        actual_value = response['SecretString']

        assert actual_value == expected_value
