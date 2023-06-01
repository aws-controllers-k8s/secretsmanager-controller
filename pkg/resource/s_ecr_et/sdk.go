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

// Code generated by ack-generate. DO NOT EDIT.

package s_ecr_et

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/secretsmanager"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/secretsmanager-controller/apis/v1alpha1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = strings.ToLower("")
	_ = &aws.JSONValue{}
	_ = &svcsdk.SecretsManager{}
	_ = &svcapitypes.Secret{}
	_ = ackv1alpha1.AWSAccountID("")
	_ = &ackerr.NotFound
	_ = &ackcondition.NotManagedMessage
	_ = &reflect.Value{}
	_ = fmt.Sprintf("")
	_ = &ackrequeue.NoRequeue{}
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkFind")
	defer func() {
		exit(err)
	}()
	// If any required fields in the input shape are missing, AWS resource is
	// not created yet. Return NotFound here to indicate to callers that the
	// resource isn't yet created.
	if rm.requiredFieldsMissingFromReadOneInput(r) {
		return nil, ackerr.NotFound
	}

	input, err := rm.newDescribeRequestPayload(r)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.DescribeSecretOutput
	resp, err = rm.sdkapi.DescribeSecretWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_ONE", "DescribeSecret", err)
	if err != nil {
		if reqErr, ok := ackerr.AWSRequestFailure(err); ok && reqErr.StatusCode() == 404 {
			return nil, ackerr.NotFound
		}
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "UNKNOWN" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.ARN != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.ARN)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.Description != nil {
		ko.Spec.Description = resp.Description
	} else {
		ko.Spec.Description = nil
	}
	if resp.KmsKeyId != nil {
		ko.Spec.KMSKeyID = resp.KmsKeyId
	} else {
		ko.Spec.KMSKeyID = nil
	}
	if resp.Name != nil {
		ko.Spec.Name = resp.Name
	} else {
		ko.Spec.Name = nil
	}
	if resp.ReplicationStatus != nil {
		f12 := []*svcapitypes.ReplicationStatusType{}
		for _, f12iter := range resp.ReplicationStatus {
			f12elem := &svcapitypes.ReplicationStatusType{}
			if f12iter.KmsKeyId != nil {
				f12elem.KMSKeyID = f12iter.KmsKeyId
			}
			if f12iter.LastAccessedDate != nil {
				f12elem.LastAccessedDate = &metav1.Time{*f12iter.LastAccessedDate}
			}
			if f12iter.Region != nil {
				f12elem.Region = f12iter.Region
			}
			if f12iter.Status != nil {
				f12elem.Status = f12iter.Status
			}
			if f12iter.StatusMessage != nil {
				f12elem.StatusMessage = f12iter.StatusMessage
			}
			f12 = append(f12, f12elem)
		}
		ko.Status.ReplicationStatus = f12
	} else {
		ko.Status.ReplicationStatus = nil
	}
	if resp.Tags != nil {
		f16 := []*svcapitypes.Tag{}
		for _, f16iter := range resp.Tags {
			f16elem := &svcapitypes.Tag{}
			if f16iter.Key != nil {
				f16elem.Key = f16iter.Key
			}
			if f16iter.Value != nil {
				f16elem.Value = f16iter.Value
			}
			f16 = append(f16, f16elem)
		}
		ko.Spec.Tags = f16
	} else {
		ko.Spec.Tags = nil
	}

	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
}

// requiredFieldsMissingFromReadOneInput returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromReadOneInput(
	r *resource,
) bool {
	return r.ko.Spec.Name == nil

}

// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.DescribeSecretInput, error) {
	res := &svcsdk.DescribeSecretInput{}

	if r.ko.Spec.Name != nil {
		res.SetSecretId(*r.ko.Spec.Name)
	}

	return res, nil
}

// sdkCreate creates the supplied resource in the backend AWS service API and
// returns a copy of the resource with resource fields (in both Spec and
// Status) filled in with values from the CREATE API operation's Output shape.
func (rm *resourceManager) sdkCreate(
	ctx context.Context,
	desired *resource,
) (created *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkCreate")
	defer func() {
		exit(err)
	}()
	input, err := rm.newCreateRequestPayload(ctx, desired)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.CreateSecretOutput
	_ = resp
	resp, err = rm.sdkapi.CreateSecretWithContext(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "CreateSecret", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.ARN != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.ARN)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.Name != nil {
		ko.Spec.Name = resp.Name
	} else {
		ko.Spec.Name = nil
	}
	if resp.ReplicationStatus != nil {
		f2 := []*svcapitypes.ReplicationStatusType{}
		for _, f2iter := range resp.ReplicationStatus {
			f2elem := &svcapitypes.ReplicationStatusType{}
			if f2iter.KmsKeyId != nil {
				f2elem.KMSKeyID = f2iter.KmsKeyId
			}
			if f2iter.LastAccessedDate != nil {
				f2elem.LastAccessedDate = &metav1.Time{*f2iter.LastAccessedDate}
			}
			if f2iter.Region != nil {
				f2elem.Region = f2iter.Region
			}
			if f2iter.Status != nil {
				f2elem.Status = f2iter.Status
			}
			if f2iter.StatusMessage != nil {
				f2elem.StatusMessage = f2iter.StatusMessage
			}
			f2 = append(f2, f2elem)
		}
		ko.Status.ReplicationStatus = f2
	} else {
		ko.Status.ReplicationStatus = nil
	}
	if resp.VersionId != nil {
		ko.Status.VersionID = resp.VersionId
	} else {
		ko.Status.VersionID = nil
	}

	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	ctx context.Context,
	r *resource,
) (*svcsdk.CreateSecretInput, error) {
	res := &svcsdk.CreateSecretInput{}

	if r.ko.Spec.AddReplicaRegions != nil {
		f0 := []*svcsdk.ReplicaRegionType{}
		for _, f0iter := range r.ko.Spec.AddReplicaRegions {
			f0elem := &svcsdk.ReplicaRegionType{}
			if f0iter.KMSKeyID != nil {
				f0elem.SetKmsKeyId(*f0iter.KMSKeyID)
			}
			if f0iter.Region != nil {
				f0elem.SetRegion(*f0iter.Region)
			}
			f0 = append(f0, f0elem)
		}
		res.SetAddReplicaRegions(f0)
	}
	if r.ko.Spec.ClientRequestToken != nil {
		res.SetClientRequestToken(*r.ko.Spec.ClientRequestToken)
	}
	if r.ko.Spec.Description != nil {
		res.SetDescription(*r.ko.Spec.Description)
	}
	if r.ko.Spec.ForceOverwriteReplicaSecret != nil {
		res.SetForceOverwriteReplicaSecret(*r.ko.Spec.ForceOverwriteReplicaSecret)
	}
	if r.ko.Spec.KMSKeyID != nil {
		res.SetKmsKeyId(*r.ko.Spec.KMSKeyID)
	}
	if r.ko.Spec.Name != nil {
		res.SetName(*r.ko.Spec.Name)
	}
	if r.ko.Spec.SecretBinary != nil {
		res.SetSecretBinary(r.ko.Spec.SecretBinary)
	}
	if r.ko.Spec.SecretString != nil {
		res.SetSecretString(*r.ko.Spec.SecretString)
	}
	if r.ko.Spec.Tags != nil {
		f8 := []*svcsdk.Tag{}
		for _, f8iter := range r.ko.Spec.Tags {
			f8elem := &svcsdk.Tag{}
			if f8iter.Key != nil {
				f8elem.SetKey(*f8iter.Key)
			}
			if f8iter.Value != nil {
				f8elem.SetValue(*f8iter.Value)
			}
			f8 = append(f8, f8elem)
		}
		res.SetTags(f8)
	}

	return res, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkUpdate")
	defer func() {
		exit(err)
	}()
	input, err := rm.newUpdateRequestPayload(ctx, desired, delta)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.UpdateSecretOutput
	_ = resp
	resp, err = rm.sdkapi.UpdateSecretWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateSecret", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.ARN != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.ARN)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.Name != nil {
		ko.Spec.Name = resp.Name
	} else {
		ko.Spec.Name = nil
	}
	if resp.VersionId != nil {
		ko.Status.VersionID = resp.VersionId
	} else {
		ko.Status.VersionID = nil
	}

	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
}

// newUpdateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Update API call for the resource
func (rm *resourceManager) newUpdateRequestPayload(
	ctx context.Context,
	r *resource,
	delta *ackcompare.Delta,
) (*svcsdk.UpdateSecretInput, error) {
	res := &svcsdk.UpdateSecretInput{}

	if r.ko.Spec.ClientRequestToken != nil {
		res.SetClientRequestToken(*r.ko.Spec.ClientRequestToken)
	}
	if r.ko.Spec.Description != nil {
		res.SetDescription(*r.ko.Spec.Description)
	}
	if r.ko.Spec.KMSKeyID != nil {
		res.SetKmsKeyId(*r.ko.Spec.KMSKeyID)
	}
	if r.ko.Spec.SecretBinary != nil {
		res.SetSecretBinary(r.ko.Spec.SecretBinary)
	}
	if r.ko.Spec.SecretString != nil {
		res.SetSecretString(*r.ko.Spec.SecretString)
	}

	return res, nil
}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkDelete")
	defer func() {
		exit(err)
	}()
	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.DeleteSecretOutput
	_ = resp
	resp, err = rm.sdkapi.DeleteSecretWithContext(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteSecret", err)
	return nil, err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeleteSecretInput, error) {
	res := &svcsdk.DeleteSecretInput{}

	return res, nil
}

// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults(
	ko *svcapitypes.Secret,
) {
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if ko.Status.ACKResourceMetadata.Region == nil {
		ko.Status.ACKResourceMetadata.Region = &rm.awsRegion
	}
	if ko.Status.ACKResourceMetadata.OwnerAccountID == nil {
		ko.Status.ACKResourceMetadata.OwnerAccountID = &rm.awsAccountID
	}
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
}

// updateConditions returns updated resource, true; if conditions were updated
// else it returns nil, false
func (rm *resourceManager) updateConditions(
	r *resource,
	onSuccess bool,
	err error,
) (*resource, bool) {
	ko := r.ko.DeepCopy()
	rm.setStatusDefaults(ko)

	// Terminal condition
	var terminalCondition *ackv1alpha1.Condition = nil
	var recoverableCondition *ackv1alpha1.Condition = nil
	var syncCondition *ackv1alpha1.Condition = nil
	for _, condition := range ko.Status.Conditions {
		if condition.Type == ackv1alpha1.ConditionTypeTerminal {
			terminalCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeRecoverable {
			recoverableCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeResourceSynced {
			syncCondition = condition
		}
	}
	var termError *ackerr.TerminalError
	if rm.terminalAWSError(err) || err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound || errors.As(err, &termError) {
		if terminalCondition == nil {
			terminalCondition = &ackv1alpha1.Condition{
				Type: ackv1alpha1.ConditionTypeTerminal,
			}
			ko.Status.Conditions = append(ko.Status.Conditions, terminalCondition)
		}
		var errorMessage = ""
		if err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound || errors.As(err, &termError) {
			errorMessage = err.Error()
		} else {
			awsErr, _ := ackerr.AWSError(err)
			errorMessage = awsErr.Error()
		}
		terminalCondition.Status = corev1.ConditionTrue
		terminalCondition.Message = &errorMessage
	} else {
		// Clear the terminal condition if no longer present
		if terminalCondition != nil {
			terminalCondition.Status = corev1.ConditionFalse
			terminalCondition.Message = nil
		}
		// Handling Recoverable Conditions
		if err != nil {
			if recoverableCondition == nil {
				// Add a new Condition containing a non-terminal error
				recoverableCondition = &ackv1alpha1.Condition{
					Type: ackv1alpha1.ConditionTypeRecoverable,
				}
				ko.Status.Conditions = append(ko.Status.Conditions, recoverableCondition)
			}
			recoverableCondition.Status = corev1.ConditionTrue
			awsErr, _ := ackerr.AWSError(err)
			errorMessage := err.Error()
			if awsErr != nil {
				errorMessage = awsErr.Error()
			}
			recoverableCondition.Message = &errorMessage
		} else if recoverableCondition != nil {
			recoverableCondition.Status = corev1.ConditionFalse
			recoverableCondition.Message = nil
		}
	}
	// Required to avoid the "declared but not used" error in the default case
	_ = syncCondition
	if terminalCondition != nil || recoverableCondition != nil || syncCondition != nil {
		return &resource{ko}, true // updated
	}
	return nil, false // not updated
}

// terminalAWSError returns awserr, true; if the supplied error is an aws Error type
// and if the exception indicates that it is a Terminal exception
// 'Terminal' exception are specified in generator configuration
func (rm *resourceManager) terminalAWSError(err error) bool {
	// No terminal_errors specified for this resource in generator config
	return false
}
