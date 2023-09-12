# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

import boto3
import pytest

from acktest import k8s

def pytest_addoption(parser):
    parser.addoption("--runslow", action="store_true", default=False, help="run slow tests")

def pytest_configure(config):
    config.addinivalue_line(
        "markers", "service(arg): mark test associated with a given service"
    )
    config.addinivalue_line(
        "markers", "slow: mark test as slow to run"
    )

def pytest_collection_modifyitems(config, items):
    if config.getoption("--runslow"):
        return
    skip_slow = pytest.mark.skip(reason="need --runslow option to run")
    for item in items:
        if "slow" in item.keywords:
            item.add_marker(skip_slow)

# Provide a k8s client to interact with the integration test cluster
@pytest.fixture(scope='class')
def k8s_client():
    return k8s._get_k8s_api_client()

@pytest.fixture(scope='module')
def secretsmanager_client():
    return boto3.client('secretsmanager')

@pytest.fixture(scope="module")
def k8s_secret():
    """Manages the lifecycle of a Kubernetes Secret for use in tests.

    Usage:
        from e2e.fixtures import k8s_secret

        class TestThing:
            def test_thing(self, k8s_secret):
                secret = k8s_secret(
                    "default", "mysecret", "mykey", "myval",
                )
    """
    created = []
    def _k8s_secret(ns, name, key, val):
        k8s.create_opaque_secret(ns, name, key, val)
        secret_ref = SecretKeyReference(ns, name, key, val)
        created.append(secret_ref)
        return secret_ref

    yield _k8s_secret

    for secret_ref in created:
        k8s.delete_secret(secret_ref.ns, secret_ref.name)
