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
	"maps"
	"slices"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	acktags "github.com/aws-controllers-k8s/runtime/pkg/tags"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	svcapitypes "github.com/aws-controllers-k8s/secretsmanager-controller/apis/v1alpha1"
)

type metricsRecorder interface {
	RecordAPICall(opType string, opID string, err error)
}

type tagsClient interface {
	TagResource(context.Context, *svcsdk.TagResourceInput, ...func(*svcsdk.Options)) (*svcsdk.TagResourceOutput, error)
	UntagResource(context.Context, *svcsdk.UntagResourceInput, ...func(*svcsdk.Options)) (*svcsdk.UntagResourceOutput, error)
}

// SyncResourceTags uses TagResource and UntagResource API Calls to add, remove
// and update resource tags.
func SyncResourceTags(
	ctx context.Context,
	client tagsClient,
	mr metricsRecorder,
	resourceARN string,
	desiredTags []*svcapitypes.Tag,
	latestTags []*svcapitypes.Tag,
) error {

	var err error
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("common.SyncResourceTags")
	defer func() {
		exit(err)
	}()

	from := toACKTags(latestTags)
	to := toACKTags(desiredTags)
	added, _, removed := ackcompare.GetTagsDifference(from, to)

	if len(removed) > 0 {
		_, err = client.UntagResource(
			ctx,
			&svcsdk.UntagResourceInput{
				SecretId: &resourceARN,
				TagKeys:  slices.Sorted(maps.Keys(removed)),
			},
		)
		mr.RecordAPICall("UPDATE", "UntagResource", err)
		if err != nil {
			return err
		}
	}

	if len(added) > 0 {
		_, err = client.TagResource(
			ctx,
			&svcsdk.TagResourceInput{
				SecretId: &resourceARN,
				Tags:     sdkTagsFromACKTags(added),
			},
		)
		mr.RecordAPICall("UPDATE", "TagResource", err)
		if err != nil {
			return err
		}
	}

	return nil
}

// toACKTags transforms a *svcapitypes.Tag array to an acktags.Tags map. Nil
// tags and tags without a key are skipped.
func toACKTags(tags []*svcapitypes.Tag) acktags.Tags {
	m := make(acktags.Tags)
	for _, t := range tags {
		if t == nil || t.Key == nil {
			continue
		}
		v := ""
		if t.Value != nil {
			v = *t.Value
		}
		m[*t.Key] = v
	}
	return m
}

// sdkTagsFromACKTags transforms an acktags.Tags map to a svcsdktypes.Tag array,
// sorted by key so that the API request is deterministic.
func sdkTagsFromACKTags(rTags acktags.Tags) []svcsdktypes.Tag {
	tags := make([]svcsdktypes.Tag, 0, len(rTags))
	for _, k := range slices.Sorted(maps.Keys(rTags)) {
		key := k // explicit copy — safe under any Go version, not just 1.22+
		val := rTags[k]
		tags = append(tags, svcsdktypes.Tag{
			Key:   &key,
			Value: &val,
		})
	}
	return tags
}
