	// we use secret name as ID for adoption, set ID when we get the ARN
	if ko.Status.ID == ko.Spec.Name {
		ko.Status.ID = resp.ARN
	}
