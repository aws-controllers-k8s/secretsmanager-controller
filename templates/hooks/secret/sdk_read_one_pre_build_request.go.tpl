	if rm.requiredFieldsMissingFromReadOneInput(r) {
		r.ko.Status.ID, err = rm.getSecretID(ctx, r)
		if err != nil {
			return nil, err
		}
	}
