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
	"fmt"
	"strings"
)

func ExtractSecretNameFromARN(arn string) string {
	// Split the ARN by ":"
	parts := strings.Split(arn, ":")

	// The secret name is the last part after "secret:"
	if len(parts) == 0 {
		return ""
	}

	secretPart := parts[len(parts)-1]
	nameVals := strings.Split(secretPart, "-")
	if len(nameVals) == 0 {
		return ""
	}

	nameHash := nameVals[len(nameVals)-1]
	return strings.TrimSuffix(secretPart, fmt.Sprintf("-%s", nameHash))
}
