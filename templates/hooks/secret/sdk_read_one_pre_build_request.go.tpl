	if rm.requiredFieldsMissingFromReadOneInput(r) {
		err = rm.attemptFindingByName(ctx, r)
        if err != nil {
            return nil, err
        }
	}
