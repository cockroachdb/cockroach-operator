package v1alpha1

func (z *AvailabilityZone) Name(base string) string {
	return base + z.StatefulSetSuffix
}
