if delta.DifferentAt("Spec.Tags") {
    err := rm.syncTags(
        ctx,
        desired,
        latest,
    )
    if err != nil {
        return nil, err
    }
}
if delta.DifferentAt("Spec.RotationEnabled") || delta.DifferentAt("Spec.RotationLambdaARN") || delta.DifferentAt("Spec.RotationRules") {
    err := rm.syncRotation(
        ctx,
        desired,
        latest,
    )
    if err != nil {
        return nil, err
    }
}
if !delta.DifferentExcept("Spec.Tags", "Spec.RotationEnabled", "Spec.RotationLambdaARN", "Spec.RotationRules") {
    return desired, nil
}
