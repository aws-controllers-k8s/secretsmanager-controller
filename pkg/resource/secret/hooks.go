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
	"context"
	"fmt"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/aws-controllers-k8s/secretsmanager-controller/pkg/resource/tags"
)

func customPreCompare(delta *ackcompare.Delta, a, b *resource) {
	compareRotationChanges(delta, a, b)
}

func compareRotationChanges(delta *ackcompare.Delta, a, b *resource) {
	// a = desired, b = latest
	// Only fire delta if the user explicitly set the field (desired != nil)
	if a.ko.Spec.RotationEnabled != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.RotationEnabled, b.ko.Spec.RotationEnabled) {
			delta.Add("Spec.RotationEnabled", a.ko.Spec.RotationEnabled, b.ko.Spec.RotationEnabled)
		} else if *a.ko.Spec.RotationEnabled != *b.ko.Spec.RotationEnabled {
			delta.Add("Spec.RotationEnabled", a.ko.Spec.RotationEnabled, b.ko.Spec.RotationEnabled)
		}
	}
	if a.ko.Spec.RotationLambdaARN != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.RotationLambdaARN, b.ko.Spec.RotationLambdaARN) {
			delta.Add("Spec.RotationLambdaARN", a.ko.Spec.RotationLambdaARN, b.ko.Spec.RotationLambdaARN)
		} else if *a.ko.Spec.RotationLambdaARN != *b.ko.Spec.RotationLambdaARN {
			delta.Add("Spec.RotationLambdaARN", a.ko.Spec.RotationLambdaARN, b.ko.Spec.RotationLambdaARN)
		}
	}
	if a.ko.Spec.RotationRules != nil {
		if b.ko.Spec.RotationRules == nil {
			delta.Add("Spec.RotationRules", a.ko.Spec.RotationRules, b.ko.Spec.RotationRules)
		} else {
			if ackcompare.HasNilDifference(a.ko.Spec.RotationRules.AutomaticallyAfterDays, b.ko.Spec.RotationRules.AutomaticallyAfterDays) ||
				(a.ko.Spec.RotationRules.AutomaticallyAfterDays != nil && *a.ko.Spec.RotationRules.AutomaticallyAfterDays != *b.ko.Spec.RotationRules.AutomaticallyAfterDays) {
				delta.Add("Spec.RotationRules", a.ko.Spec.RotationRules, b.ko.Spec.RotationRules)
			} else if ackcompare.HasNilDifference(a.ko.Spec.RotationRules.Duration, b.ko.Spec.RotationRules.Duration) ||
				(a.ko.Spec.RotationRules.Duration != nil && *a.ko.Spec.RotationRules.Duration != *b.ko.Spec.RotationRules.Duration) {
				delta.Add("Spec.RotationRules", a.ko.Spec.RotationRules, b.ko.Spec.RotationRules)
			} else if ackcompare.HasNilDifference(a.ko.Spec.RotationRules.ScheduleExpression, b.ko.Spec.RotationRules.ScheduleExpression) ||
				(a.ko.Spec.RotationRules.ScheduleExpression != nil && *a.ko.Spec.RotationRules.ScheduleExpression != *b.ko.Spec.RotationRules.ScheduleExpression) {
				delta.Add("Spec.RotationRules", a.ko.Spec.RotationRules, b.ko.Spec.RotationRules)
			}
		}
	}
}

func (rm *resourceManager) syncRotation(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncRotation")
	defer func() {
		exit(err)
	}()

	secretID := latest.ko.Status.ID
	if secretID == nil {
		return fmt.Errorf("secret ARN not available")
	}

	desiredEnabled := desired.ko.Spec.RotationEnabled != nil && *desired.ko.Spec.RotationEnabled
	latestEnabled := latest.ko.Spec.RotationEnabled != nil && *latest.ko.Spec.RotationEnabled

	if !desiredEnabled && latestEnabled {
		_, err = rm.sdkapi.CancelRotateSecret(ctx, &svcsdk.CancelRotateSecretInput{
			SecretId: secretID,
		})
		rm.metrics.RecordAPICall("UPDATE", "CancelRotateSecret", err)
		return err
	}

	if desiredEnabled {
		rotateImmediately := false
		input := &svcsdk.RotateSecretInput{
			SecretId:          secretID,
			RotateImmediately: &rotateImmediately,
		}
		if desired.ko.Spec.RotationLambdaARN != nil {
			input.RotationLambdaARN = desired.ko.Spec.RotationLambdaARN
		}
		if desired.ko.Spec.RotationRules != nil {
			input.RotationRules = &svcsdktypes.RotationRulesType{}
			if desired.ko.Spec.RotationRules.AutomaticallyAfterDays != nil {
				input.RotationRules.AutomaticallyAfterDays = desired.ko.Spec.RotationRules.AutomaticallyAfterDays
			}
			if desired.ko.Spec.RotationRules.Duration != nil {
				input.RotationRules.Duration = desired.ko.Spec.RotationRules.Duration
			}
			if desired.ko.Spec.RotationRules.ScheduleExpression != nil {
				input.RotationRules.ScheduleExpression = desired.ko.Spec.RotationRules.ScheduleExpression
			}
		}
		_, err = rm.sdkapi.RotateSecret(ctx, input)
		rm.metrics.RecordAPICall("UPDATE", "RotateSecret", err)
		return err
	}

	return nil
}

// syncTags keeps the resource's tags in sync.
func (rm *resourceManager) syncTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	return tags.SyncResourceTags(
		ctx,
		rm.sdkapi,
		rm.metrics,
		string(*latest.ko.Status.ACKResourceMetadata.ARN),
		desired.ko.Spec.Tags,
		latest.ko.Spec.Tags,
	)
}

func (rm *resourceManager) getSecretID(
	ctx context.Context,
	r *resource,
) (id *string, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.getSecretID")
	defer func() {
		exit(err)
	}()

	// although spec.name is a required field, during `adopt`,
	// user might provide an empty name and only populate the status.id
	if r.ko.Spec.Name == nil {
		return nil, nil
	}

	input := &svcsdk.ListSecretsInput{
		Filters: []svcsdktypes.Filter{{
			Key:    "name",
			Values: []string{*r.ko.Spec.Name},
		}},
	}

	for {
		resp, err := rm.sdkapi.ListSecrets(ctx, input)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, nil
		}
		for _, s := range resp.SecretList {
			if s.Name != nil && *s.Name == *r.ko.Spec.Name {
				return s.ARN, nil
			}
		}
		if resp.NextToken == nil || *resp.NextToken == "" {
			return nil, nil
		}
		input.NextToken = resp.NextToken
	}
}
