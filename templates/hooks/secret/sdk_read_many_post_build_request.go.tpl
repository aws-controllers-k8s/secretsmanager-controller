	input.Filters = []svcsdktypes.Filter{
		{
			Key: "name",
			Values: []string{*r.ko.Spec.Name},
		},
	}
