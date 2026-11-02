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

package tags

import (
	"context"
	"reflect"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	svcapitypes "github.com/aws-controllers-k8s/secretsmanager-controller/apis/v1alpha1"
)

const testARN = "arn:aws:secretsmanager:us-west-2:123456789012:secret:test"

// fakeTagsClient records the TagResource and UntagResource calls it receives.
type fakeTagsClient struct {
	tagInputs   []*svcsdk.TagResourceInput
	untagInputs []*svcsdk.UntagResourceInput
}

func (f *fakeTagsClient) TagResource(
	_ context.Context,
	in *svcsdk.TagResourceInput,
	_ ...func(*svcsdk.Options),
) (*svcsdk.TagResourceOutput, error) {
	f.tagInputs = append(f.tagInputs, in)
	return &svcsdk.TagResourceOutput{}, nil
}

func (f *fakeTagsClient) UntagResource(
	_ context.Context,
	in *svcsdk.UntagResourceInput,
	_ ...func(*svcsdk.Options),
) (*svcsdk.UntagResourceOutput, error) {
	f.untagInputs = append(f.untagInputs, in)
	return &svcsdk.UntagResourceOutput{}, nil
}

// fakeMetricsRecorder counts the API calls recorded against it.
type fakeMetricsRecorder struct {
	calls int
}

func (f *fakeMetricsRecorder) RecordAPICall(_ string, _ string, _ error) {
	f.calls++
}

func tag(key, value string) *svcapitypes.Tag {
	return &svcapitypes.Tag{Key: &key, Value: &value}
}

func TestSyncResourceTags(t *testing.T) {
	tests := []struct {
		name         string
		desired      []*svcapitypes.Tag
		latest       []*svcapitypes.Tag
		wantTags     []svcsdktypes.Tag
		wantUntags   []string
		wantTagCalls int
	}{
		{
			name:    "no-op when desired equals latest",
			desired: []*svcapitypes.Tag{tag("env", "prod")},
			latest:  []*svcapitypes.Tag{tag("env", "prod")},
		},
		{
			name:         "new tag is added with TagResource",
			desired:      []*svcapitypes.Tag{tag("env", "prod"), tag("team", "core")},
			latest:       []*svcapitypes.Tag{tag("env", "prod")},
			wantTags:     []svcsdktypes.Tag{{Key: ptrTo("team"), Value: ptrTo("core")}},
			wantTagCalls: 1,
		},
		{
			name:       "removed tag is deleted with UntagResource",
			desired:    []*svcapitypes.Tag{tag("env", "prod")},
			latest:     []*svcapitypes.Tag{tag("env", "prod"), tag("team", "core")},
			wantUntags: []string{"team"},
		},
		{
			// Tag APIs upsert, so a changed value is only sent to TagResource.
			name:         "changed tag value is updated with TagResource only",
			desired:      []*svcapitypes.Tag{tag("env", "staging")},
			latest:       []*svcapitypes.Tag{tag("env", "prod")},
			wantTags:     []svcsdktypes.Tag{{Key: ptrTo("env"), Value: ptrTo("staging")}},
			wantTagCalls: 1,
		},
		{
			name:         "tag and untag are both issued in one SyncResourceTags invocation",
			desired:      []*svcapitypes.Tag{tag("env", "prod"), tag("new", "val")},
			latest:       []*svcapitypes.Tag{tag("env", "prod"), tag("old", "x")},
			wantTags:     []svcsdktypes.Tag{{Key: ptrTo("new"), Value: ptrTo("val")}},
			wantUntags:   []string{"old"},
			wantTagCalls: 1,
		},
		{
			name:         "nil tag in desired slice does not panic",
			desired:      []*svcapitypes.Tag{nil, tag("env", "prod")},
			latest:       []*svcapitypes.Tag{nil},
			wantTags:     []svcsdktypes.Tag{{Key: ptrTo("env"), Value: ptrTo("prod")}},
			wantTagCalls: 1,
		},
		{
			// toACKTags maps a nil Value to "", so the SDK tag carries an empty string.
			name:         "tag with nil value is sent with empty string value",
			desired:      []*svcapitypes.Tag{{Key: ptrTo("owner")}},
			latest:       []*svcapitypes.Tag{},
			wantTags:     []svcsdktypes.Tag{{Key: ptrTo("owner"), Value: ptrTo("")}},
			wantTagCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &fakeTagsClient{}
			mr := &fakeMetricsRecorder{}

			err := SyncResourceTags(context.Background(), client, mr, testARN, test.desired, test.latest)
			if err != nil {
				t.Fatalf("SyncResourceTags() returned unexpected error: %v", err)
			}

			if len(client.tagInputs) != test.wantTagCalls {
				t.Fatalf("TagResource calls = %d, want %d", len(client.tagInputs), test.wantTagCalls)
			}
			if test.wantTagCalls > 0 {
				got := client.tagInputs[0]
				if got.SecretId == nil || *got.SecretId != testARN {
					t.Errorf("TagResource SecretId = %v, want %q", got.SecretId, testARN)
				}
				if !reflect.DeepEqual(got.Tags, test.wantTags) {
					t.Errorf("TagResource Tags = %+v, want %+v", derefTags(got.Tags), derefTags(test.wantTags))
				}
			}

			wantUntagCalls := 0
			if len(test.wantUntags) > 0 {
				wantUntagCalls = 1
			}
			if len(client.untagInputs) != wantUntagCalls {
				t.Fatalf("UntagResource calls = %d, want %d", len(client.untagInputs), wantUntagCalls)
			}
			if wantUntagCalls == 1 {
				got := client.untagInputs[0]
				if got.SecretId == nil || *got.SecretId != testARN {
					t.Errorf("UntagResource SecretId = %v, want %q", got.SecretId, testARN)
				}
				if !reflect.DeepEqual(got.TagKeys, test.wantUntags) {
					t.Errorf("UntagResource TagKeys = %v, want %v", got.TagKeys, test.wantUntags)
				}
			}

			if wantCalls := test.wantTagCalls + wantUntagCalls; mr.calls != wantCalls {
				t.Errorf("RecordAPICall count = %d, want %d", mr.calls, wantCalls)
			}
		})
	}
}

func ptrTo[T any](v T) *T {
	return &v
}

func derefTags(tags []svcsdktypes.Tag) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		k, v := "<nil>", "<nil>"
		if t.Key != nil {
			k = *t.Key
		}
		if t.Value != nil {
			v = *t.Value
		}
		out = append(out, k+"="+v)
	}
	return out
}
