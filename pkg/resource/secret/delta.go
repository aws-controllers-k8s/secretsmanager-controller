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

package secret

import (
	"bytes"
	"reflect"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	acktags "github.com/aws-controllers-k8s/runtime/pkg/tags"
)

// Hack to avoid import errors during build...
var (
	_ = &bytes.Buffer{}
	_ = &reflect.Method{}
	_ = &acktags.Tags{}
)

// newResourceDelta returns a new `ackcompare.Delta` used to compare two
// resources
func newResourceDelta(
	a *resource,
	b *resource,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if (a == nil && b != nil) ||
		(a != nil && b == nil) {
		delta.Add("", a, b)
		return delta
	}

	if ackcompare.HasNilDifference(a.ko.Spec.Description, b.ko.Spec.Description) {
		delta.Add("Spec.Description", a.ko.Spec.Description, b.ko.Spec.Description)
	} else if a.ko.Spec.Description != nil && b.ko.Spec.Description != nil {
		if *a.ko.Spec.Description != *b.ko.Spec.Description {
			delta.Add("Spec.Description", a.ko.Spec.Description, b.ko.Spec.Description)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.ForceOverwriteReplicaSecret, b.ko.Spec.ForceOverwriteReplicaSecret) {
		delta.Add("Spec.ForceOverwriteReplicaSecret", a.ko.Spec.ForceOverwriteReplicaSecret, b.ko.Spec.ForceOverwriteReplicaSecret)
	} else if a.ko.Spec.ForceOverwriteReplicaSecret != nil && b.ko.Spec.ForceOverwriteReplicaSecret != nil {
		if *a.ko.Spec.ForceOverwriteReplicaSecret != *b.ko.Spec.ForceOverwriteReplicaSecret {
			delta.Add("Spec.ForceOverwriteReplicaSecret", a.ko.Spec.ForceOverwriteReplicaSecret, b.ko.Spec.ForceOverwriteReplicaSecret)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.KMSKeyID, b.ko.Spec.KMSKeyID) {
		delta.Add("Spec.KMSKeyID", a.ko.Spec.KMSKeyID, b.ko.Spec.KMSKeyID)
	} else if a.ko.Spec.KMSKeyID != nil && b.ko.Spec.KMSKeyID != nil {
		if *a.ko.Spec.KMSKeyID != *b.ko.Spec.KMSKeyID {
			delta.Add("Spec.KMSKeyID", a.ko.Spec.KMSKeyID, b.ko.Spec.KMSKeyID)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.KMSKeyRef, b.ko.Spec.KMSKeyRef) {
		delta.Add("Spec.KMSKeyRef", a.ko.Spec.KMSKeyRef, b.ko.Spec.KMSKeyRef)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Name, b.ko.Spec.Name) {
		delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
	} else if a.ko.Spec.Name != nil && b.ko.Spec.Name != nil {
		if *a.ko.Spec.Name != *b.ko.Spec.Name {
			delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
		}
	}
	if len(a.ko.Spec.ReplicaRegions) != len(b.ko.Spec.ReplicaRegions) {
		delta.Add("Spec.ReplicaRegions", a.ko.Spec.ReplicaRegions, b.ko.Spec.ReplicaRegions)
	} else if len(a.ko.Spec.ReplicaRegions) > 0 {
		if !reflect.DeepEqual(a.ko.Spec.ReplicaRegions, b.ko.Spec.ReplicaRegions) {
			delta.Add("Spec.ReplicaRegions", a.ko.Spec.ReplicaRegions, b.ko.Spec.ReplicaRegions)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.SecretString, b.ko.Spec.SecretString) {
		delta.Add("Spec.SecretString", a.ko.Spec.SecretString, b.ko.Spec.SecretString)
	} else if a.ko.Spec.SecretString != nil && b.ko.Spec.SecretString != nil {
		if *a.ko.Spec.SecretString != *b.ko.Spec.SecretString {
			delta.Add("Spec.SecretString", a.ko.Spec.SecretString, b.ko.Spec.SecretString)
		}
	}
	desiredACKTags, _ := convertToOrderedACKTags(a.ko.Spec.Tags)
	latestACKTags, _ := convertToOrderedACKTags(b.ko.Spec.Tags)
	if !ackcompare.MapStringStringEqual(desiredACKTags, latestACKTags) {
		delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
	}

	return delta
}
