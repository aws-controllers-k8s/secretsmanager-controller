// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package secret

import (
	"testing"

	svcsdk "github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	svcapitypes "github.com/aws-controllers-k8s/secretsmanager-controller/apis/v1alpha1"
)

func secretWithRecoveryWindow(recoveryWindowInDays *int64) *resource {
	return &resource{
		ko: &svcapitypes.Secret{
			Spec: svcapitypes.SecretSpec{
				RecoveryWindowInDays: recoveryWindowInDays,
			},
		},
	}
}

func TestSetDeleteSecretInput(t *testing.T) {
	tests := []struct {
		name string
		// recoveryWindowInDays is the spec value; nil means unset.
		recoveryWindowInDays *int64
		wantForceDelete      *bool
		wantRecoveryWindow   *int64
	}{
		{
			name: "unset leaves the service default of 30 days",
		},
		{
			// Secrets Manager has no zero day recovery window, so a zero
			// request becomes the mutually exclusive force delete parameter.
			name:                 "zero forces deletion without recovery",
			recoveryWindowInDays: ptrTo(int64(0)),
			wantForceDelete:      ptrTo(true),
		},
		{
			name:                 "non zero window is passed through",
			recoveryWindowInDays: ptrTo(int64(7)),
			wantRecoveryWindow:   ptrTo(int64(7)),
		},
		{
			// Secrets Manager owns the accepted range, so an out of range
			// value reaches the API and is rejected there rather than being
			// caught by the controller.
			name:                 "out of range window is still passed through",
			recoveryWindowInDays: ptrTo(int64(31)),
			wantRecoveryWindow:   ptrTo(int64(31)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := secretWithRecoveryWindow(test.recoveryWindowInDays)
			// The generated payload copies the spec value verbatim, which is
			// the state this hook is given.
			input := &svcsdk.DeleteSecretInput{
				RecoveryWindowInDays: test.recoveryWindowInDays,
			}

			setDeleteSecretInput(r, input)

			assertBoolPtr(t, "input.ForceDeleteWithoutRecovery", input.ForceDeleteWithoutRecovery, test.wantForceDelete)
			assertInt64Ptr(t, "input.RecoveryWindowInDays", input.RecoveryWindowInDays, test.wantRecoveryWindow)
		})
	}
}

func TestSetDeleteSecretInput_PreservesSecretID(t *testing.T) {
	secretID := "arn:aws:secretsmanager:us-west-2:123456789012:secret:my-secret-AbCdEf"
	zero := int64(0)
	r := secretWithRecoveryWindow(&zero)
	input := &svcsdk.DeleteSecretInput{
		SecretId:             &secretID,
		RecoveryWindowInDays: &zero,
	}

	setDeleteSecretInput(r, input)

	if input.SecretId == nil {
		t.Fatal("input.SecretId = nil, want it preserved")
	}
	if *input.SecretId != secretID {
		t.Errorf("input.SecretId = %q, want %q", *input.SecretId, secretID)
	}
}

func ptrTo[T any](v T) *T {
	return &v
}

func assertBoolPtr(t *testing.T, name string, got, want *bool) {
	t.Helper()
	switch {
	case want == nil && got != nil:
		t.Errorf("%s = %t, want nil", name, *got)
	case want != nil && got == nil:
		t.Errorf("%s = nil, want %t", name, *want)
	case want != nil && got != nil && *got != *want:
		t.Errorf("%s = %t, want %t", name, *got, *want)
	}
}

func assertInt64Ptr(t *testing.T, name string, got, want *int64) {
	t.Helper()
	switch {
	case want == nil && got != nil:
		t.Errorf("%s = %d, want nil", name, *got)
	case want != nil && got == nil:
		t.Errorf("%s = nil, want %d", name, *want)
	case want != nil && got != nil && *got != *want:
		t.Errorf("%s = %d, want %d", name, *got, *want)
	}
}
