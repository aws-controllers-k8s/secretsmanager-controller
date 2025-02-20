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

package v1alpha1

import (
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretSpec defines the desired state of Secret.
type SecretSpec struct {

	// The description of the secret.
	Description *string `json:"description,omitempty"`
	// Specifies whether to overwrite a secret with the same name in the destination
	// Region. By default, secrets aren't overwritten.
	ForceOverwriteReplicaSecret *bool `json:"forceOverwriteReplicaSecret,omitempty"`
	// The ARN, key ID, or alias of the KMS key that Secrets Manager uses to encrypt
	// the secret value in the secret. An alias is always prefixed by alias/, for
	// example alias/aws/secretsmanager. For more information, see About aliases
	// (https://docs.aws.amazon.com/kms/latest/developerguide/alias-about.html).
	//
	// To use a KMS key in a different account, use the key ARN or the alias ARN.
	//
	// If you don't specify this value, then Secrets Manager uses the key aws/secretsmanager.
	// If that key doesn't yet exist, then Secrets Manager creates it for you automatically
	// the first time it encrypts the secret value.
	//
	// If the secret is in a different Amazon Web Services account from the credentials
	// calling the API, then you can't use aws/secretsmanager to encrypt the secret,
	// and you must create and use a customer managed KMS key.
	KMSKeyID *string `json:"kmsKeyID,omitempty"`
	// The name of the new secret.
	//
	// The secret name can contain ASCII letters, numbers, and the following characters:
	// /_+=.@-
	//
	// Do not end your secret name with a hyphen followed by six characters. If
	// you do so, you risk confusion and unexpected results when searching for a
	// secret by partial ARN. Secrets Manager automatically adds a hyphen and six
	// random characters after the secret name at the end of the ARN.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable once set"
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// A list of Regions and KMS keys to replicate secrets.
	ReplicaRegions []*ReplicaRegionType `json:"replicaRegions,omitempty"`
	// The text data to encrypt and store in this new version of the secret. We
	// recommend you use a JSON structure of key/value pairs for your secret value.
	//
	// Either SecretString or SecretBinary must have a value, but not both.
	//
	// If you create a secret by using the Secrets Manager console then Secrets
	// Manager puts the protected secret text in only the SecretString parameter.
	// The Secrets Manager console stores the information as a JSON structure of
	// key/value pairs that a Lambda rotation function can parse.
	//
	// Sensitive: This field contains sensitive information, so the service does
	// not include it in CloudTrail log entries. If you create your own log entries,
	// you must also avoid logging the information in this field.
	SecretString *ackv1alpha1.SecretKeyReference `json:"secretString,omitempty"`
	// A list of tags to attach to the secret. Each tag is a key and value pair
	// of strings in a JSON text string, for example:
	//
	// [{"Key":"CostCenter","Value":"12345"},{"Key":"environment","Value":"production"}]
	//
	// Secrets Manager tag key names are case sensitive. A tag with the key "ABC"
	// is a different tag from one with key "abc".
	//
	// If you check tags in permissions policies as part of your security strategy,
	// then adding or removing a tag can change permissions. If the completion of
	// this operation would result in you losing your permissions for this secret,
	// then Secrets Manager blocks the operation and returns an Access Denied error.
	// For more information, see Control access to secrets using tags (https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_examples.html#tag-secrets-abac)
	// and Limit access to identities with tags that match secrets' tags (https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_examples.html#auth-and-access_tags2).
	//
	// For information about how to format a JSON parameter for the various command
	// line tool environments, see Using JSON for Parameters (https://docs.aws.amazon.com/cli/latest/userguide/cli-using-param.html#cli-using-param-json).
	// If your command-line tool or SDK requires quotation marks around the parameter,
	// you should use single quotes to avoid confusion with the double quotes required
	// in the JSON text.
	//
	// For tag quotas and naming restrictions, see Service quotas for Tagging (https://docs.aws.amazon.com/general/latest/gr/arg.html#taged-reference-quotas)
	// in the Amazon Web Services General Reference guide.
	Tags []*Tag `json:"tags,omitempty"`
}

// SecretStatus defines the observed state of Secret
type SecretStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	// +kubebuilder:validation:Optional
	ACKResourceMetadata *ackv1alpha1.ResourceMetadata `json:"ackResourceMetadata"`
	// All CRs managed by ACK have a common `Status.Conditions` member that
	// contains a collection of `ackv1alpha1.Condition` objects that describe
	// the various terminal states of the CR and its backend AWS service API
	// resource
	// +kubebuilder:validation:Optional
	Conditions []*ackv1alpha1.Condition `json:"conditions"`
	// The ARN of the secret.
	// +kubebuilder:validation:Optional
	ID *string `json:"id,omitempty"`
	// A list of the replicas of this secret and their status:
	//
	//    * Failed, which indicates that the replica was not created.
	//
	//    * InProgress, which indicates that Secrets Manager is in the process of
	//    creating the replica.
	//
	//    * InSync, which indicates that the replica was created.
	// +kubebuilder:validation:Optional
	ReplicationStatus []*ReplicationStatusType `json:"replicationStatus,omitempty"`
	// The unique identifier associated with the version of the new secret.
	// +kubebuilder:validation:Optional
	VersionID *string `json:"versionID,omitempty"`
}

// Secret is the Schema for the Secrets API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Secret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretSpec   `json:"spec,omitempty"`
	Status            SecretStatus `json:"status,omitempty"`
}

// SecretList contains a list of Secret
// +kubebuilder:object:root=true
type SecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Secret{}, &SecretList{})
}
