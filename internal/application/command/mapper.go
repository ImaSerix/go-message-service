package command

func idConverter[From ~[16]byte, To ~[16]byte](v From) To {
	return To(v)
}

func idSliceConverter[From ~[16]byte, To ~[16]byte](v []From) []To {
	res := make([]To, 0, len(v))
	for _, e := range v {
		res = append(res, idConverter[From, To](e))
	}
	return res
}
