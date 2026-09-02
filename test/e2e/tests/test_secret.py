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

CREATE_WAIT_AFTER_SECONDS = 5
DELETE_WAIT_AFTER_SECONDS = 5
UPDATE_WAIT_AFTER_SECONDS = 5

# Secrets Manager purges a force deleted secret in a background process with no
# timing guarantee. Observed at roughly ten seconds, so this leaves headroom
# while staying far below the 7 day minimum recovery window a windowed delete
# would hold the secret for.
FORCE_DELETE_PURGE_TIMEOUT_SECONDS = 60
FORCE_DELETE_POLL_INTERVAL_SECONDS = 5


@pytest.fixture(scope="module")
def simple_secret(
        request,
        secretsmanager_client,
        k8s_secret,
):
    secret_str_ns = "default"
    secret_str_name = request.param.get("name")
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
    @pytest.mark.parametrize(
        "simple_secret",
        [
            {
                "name": "test-secret",
            },
        ],
        indirect=True,
    )
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

    def test_delete_without_recovery_window(self, secretsmanager_client, k8s_secret):
        """A zero recovery window force deletes the secret, so Secrets Manager
        schedules the purge immediately instead of holding it for 30 days.
        """
        secret = k8s_secret(
            "default", random_suffix_name("no-recovery-str", 24),
            "secret_str_key", '{"env":"test"}',
        )
        resource_name = random_suffix_name("no-recovery-secret", 24)

        replacements = REPLACEMENT_VALUES.copy()
        replacements["SECRET_NAME"] = resource_name
        replacements["K8S_SECRET_NAMESPACE"] = secret.ns
        replacements["K8S_SECRET_NAME"] = secret.name
        replacements["K8S_SECRET_KEY"] = secret.key

        resource_data = load_secretsmanager_resource(
            "secret",
            additional_replacements=replacements,
        )
        resource_data["spec"]["recoveryWindowInDays"] = 0

        ref = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
            resource_name, namespace="default",
        )

        k8s.create_custom_resource(ref, resource_data)
        k8s.wait_resource_consumed_by_controller(ref)
        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'arn' in cr['status']['ackResourceMetadata']

        _, deleted = k8s.delete_custom_resource(
            ref,
            period_length=DELETE_WAIT_AFTER_SECONDS,
        )
        assert deleted

        self._assert_deleted_without_recovery_window(
            secretsmanager_client, resource_name,
        )

    def _assert_deleted_without_recovery_window(
        self, secretsmanager_client, secret_name,
    ):
        # DescribeSecret reports DeletedDate as the time DeleteSecret was called,
        # not the scheduled purge, so it reads the same for a windowed delete as
        # for a forced one. Being purged outright is what distinguishes a force
        # delete, so poll until the secret is gone: a windowed delete would keep
        # it describable for at least the 7 day minimum window.
        deadline = time.monotonic() + FORCE_DELETE_PURGE_TIMEOUT_SECONDS
        while time.monotonic() < deadline:
            try:
                secretsmanager_client.describe_secret(SecretId=secret_name)
            except secretsmanager_client.exceptions.ResourceNotFoundException:
                return
            time.sleep(FORCE_DELETE_POLL_INTERVAL_SECONDS)

        pytest.fail(
            f"secret {secret_name} still exists "
            f"{FORCE_DELETE_PURGE_TIMEOUT_SECONDS}s after deletion, so a "
            "recovery window may have been applied or the secret was not force deleted"
        )

    @pytest.mark.parametrize(
        "simple_secret",
        [
            {
                "name": "tag-secret",
            },
        ],
        indirect=True,
    )
    def test_tag_update(self, secretsmanager_client, simple_secret):
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

        # Update the tags
        new_tags = [
            {"key": "key1", "value": "new_value_1"},
            {"key": "key2", "value": "value2"},
        ]
        cr["spec"]["tags"] = new_tags

        k8s.patch_custom_resource(ref, cr)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)
        

        # Check that the tags were updated
        expect_tags = {
            "key1": "new_value_1",
            "key2": "value2"
        }
        secretsmanager_validator.assert_tags(secret_name, expect_tags)
