	// For adoption, we need to check if the secret exists in AWS Secrets Manager, but not in our k8s CR.
	if r.ko.Status.ID == nil{
		r.ko.Status.ID = r.ko.Spec.Name
	}
