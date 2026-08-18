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
	"errors"
	"testing"

	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
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
			name:                 "minimum recovery window is passed through",
			recoveryWindowInDays: ptrTo(int64(7)),
			wantRecoveryWindow:   ptrTo(int64(7)),
		},
		{
			name:                 "maximum recovery window is passed through",
			recoveryWindowInDays: ptrTo(int64(30)),
			wantRecoveryWindow:   ptrTo(int64(30)),
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

			if err := setDeleteSecretInput(r, input); err != nil {
				t.Fatalf("setDeleteSecretInput() returned unexpected error: %v", err)
			}

			assertBoolPtr(t, "input.ForceDeleteWithoutRecovery", input.ForceDeleteWithoutRecovery, test.wantForceDelete)
			assertInt64Ptr(t, "input.RecoveryWindowInDays", input.RecoveryWindowInDays, test.wantRecoveryWindow)
		})
	}
}

// Secrets Manager accepts 0 or 7 through 30. Anything else would be rejected on
// every reconcile, so it fails terminally instead of looping.
func TestSetDeleteSecretInput_OutOfRangeIsTerminal(t *testing.T) {
	tests := []struct {
		name                 string
		recoveryWindowInDays int64
	}{
		{name: "between zero and the minimum", recoveryWindowInDays: 6},
		{name: "just above the maximum", recoveryWindowInDays: 31},
		{name: "negative", recoveryWindowInDays: -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := secretWithRecoveryWindow(&test.recoveryWindowInDays)
			input := &svcsdk.DeleteSecretInput{
				RecoveryWindowInDays: &test.recoveryWindowInDays,
			}

			err := setDeleteSecretInput(r, input)
			if err == nil {
				t.Fatalf("setDeleteSecretInput() with recoveryWindowInDays=%d returned nil error, want an error", test.recoveryWindowInDays)
			}
			var terminalErr *ackerr.TerminalError
			if !errors.As(err, &terminalErr) {
				t.Errorf("setDeleteSecretInput() returned %T, want a *ackerr.TerminalError", err)
			}
		})
	}
}

// The two parameters are mutually exclusive; sending both fails the API call.
func TestSetDeleteSecretInput_ZeroDoesNotSendBothParameters(t *testing.T) {
	zero := int64(0)
	r := secretWithRecoveryWindow(&zero)
	input := &svcsdk.DeleteSecretInput{RecoveryWindowInDays: &zero}

	if err := setDeleteSecretInput(r, input); err != nil {
		t.Fatalf("setDeleteSecretInput() returned unexpected error: %v", err)
	}

	if input.ForceDeleteWithoutRecovery == nil {
		t.Fatal("input.ForceDeleteWithoutRecovery = nil, want true")
	}
	if input.RecoveryWindowInDays != nil {
		t.Errorf("input.RecoveryWindowInDays = %d, want nil so it is omitted from the request", *input.RecoveryWindowInDays)
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

	if err := setDeleteSecretInput(r, input); err != nil {
		t.Fatalf("setDeleteSecretInput() returned unexpected error: %v", err)
	}

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
