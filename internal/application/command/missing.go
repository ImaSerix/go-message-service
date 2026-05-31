package command

// result is slice with elements, which has a, but does not have b
func missing[T ~[16]byte](a, b []T) []T {
	mb := make(map[T]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var miss []T
	for _, x := range a {
		if _, found := mb[x]; !found {
			miss = append(miss, x)
		}
	}
	return miss
}
