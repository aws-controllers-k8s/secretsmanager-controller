    tmp, ok := identifier.AdditionalKeys["id"]
	if !ok {
		return ackerrors.NewTerminalError(fmt.Errorf("required field missing: id"))
	}
	r.ko.Status.ID = &tmp
