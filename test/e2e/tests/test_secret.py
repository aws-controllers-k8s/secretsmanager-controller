# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# 	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for the SecretsManager Secret API.
"""

import logging
import pytest
import time
from e2e.fixtures import k8s_secret
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_secretsmanager_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import SecretsManagerValidator

RESOURCE_KIND = "Secret"
RESOURCE_PLURAL = "secrets"

DELETE_WAIT_AFTER_SECONDS = 5


@pytest.fixture(scope="module")
def simple_secret(
        secretsmanager_client,
        k8s_secret,
):
    secret_str_ns = "default"
    secret_str_name = "secret-name"
    secret_str_key = "secret_str_key"
    secret_str_val = '{"env":"test"}'

    secret = k8s_secret(
        secret_str_ns,
        secret_str_name,
        secret_str_key,
        secret_str_val,
    )
    resource_name = random_suffix_name("simple-secret", 24)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["SECRET_NAME"] = resource_name
    replacements["K8S_SECRET_NAMESPACE"] = secret.ns
    replacements["K8S_SECRET_NAME"] = secret.name
    replacements["K8S_SECRET_KEY"] = secret.key

    # Load resource
    resource_data = load_secretsmanager_resource(
        "secret",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        resource_name, namespace="default",
    )

    # Create secret
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    yield cr, ref

    # Delete k8s resource
    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    response = secretsmanager_client.describe_secret(SecretId=resource_name)
    delete_date = response['DeletedDate']
    assert delete_date is not None


@service_marker
class TestSecret:
    def test_create_delete(self, secretsmanager_client, simple_secret):
        (res, ref) = simple_secret

        time.sleep(5)

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'name' in cr["spec"]
        assert 'arn' in cr['status']['ackResourceMetadata']

        secret_name = cr['spec']['name']
        secretsmanager_validator = SecretsManagerValidator(secretsmanager_client)
        expect_tags = {
            "key1": "value1",
        }
        secretsmanager_validator.assert_tags(secret_name, expect_tags)

        expected_value = '{"env":"test"}'
        secretsmanager_validator.assert_secret_value(secret_name, expected_value)
