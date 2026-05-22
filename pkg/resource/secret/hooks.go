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

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	svcapitypes "github.com/aws-controllers-k8s/secretsmanager-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/secretsmanager-controller/pkg/resource/tags"
)

func customPreCompare(delta *ackcompare.Delta, a, b *resource) {
	compareSecretReferenceChanges(delta, a, b)
}

func getLastAppliedSecretReferenceString(r *ackv1alpha1.SecretKeyReference) string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s.%s", r.Namespace, r.Name, r.Key)
}

func setLastAppliedSecretReferenceAnnotation(r *resource) {
	if r.ko.Annotations == nil {
		r.ko.Annotations = make(map[string]string)
	}
	r.ko.Annotations[svcapitypes.LastAppliedSecretAnnotation] = getLastAppliedSecretReferenceString(r.ko.Spec.SecretString)
}

func getLastAppliedSecretReferenceAnnotation(r *resource) string {
	if r.ko.Annotations == nil {
		return ""
	}
	return r.ko.Annotations[svcapitypes.LastAppliedSecretAnnotation]
}

func compareSecretReferenceChanges(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	oldRef := getLastAppliedSecretReferenceAnnotation(a)
	newRef := getLastAppliedSecretReferenceString(a.ko.Spec.SecretString)
	if oldRef != newRef {
		delta.Add("Spec.SecretString", oldRef, newRef)
	}
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
