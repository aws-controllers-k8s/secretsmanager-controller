	if input.SecretId == nil {
		input.SecretId = r.ko.Spec.Name
	}
