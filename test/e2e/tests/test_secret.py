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

import pytest
import time
import logging
import boto3

from acktest.resources import random_suffix_name
from acktest.k8s import resource as k8s
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_secretsmanager_resource
from e2e.replacement_values import REPLACEMENT_VALUES

RESOURCE_KIND = "Secret"
RESOURCE_PLURAL = "secrets"


@pytest.fixture(scope="module")
def simple_secret():
    resource_name = random_suffix_name("simple-secret", 24)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["SECRET_NAME"] = resource_name

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

    DELETE_WAIT_AFTER_SECONDS = 5
    # Delete k8s resource
    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted


@service_marker
@pytest.mark.canary
class TestSecret:
    def test_create(self, simple_secret):
        (res, ref) = simple_secret

        time.sleep(5)

        logging.debug(ref)
        logging.debug(res)
        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'arn' in cr['status']['ackResourceMetadata']
