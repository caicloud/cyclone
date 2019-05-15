package utils

// MergeMap merges map 'in' into map 'out', if there is an element in both 'in' and 'out',
// we will use the value of 'in'.
func MergeMap(in, out map[string]string) map[string]string {
	if in != nil {
		if out == nil {
			out = make(map[string]string)
		}

		for k, v := range in {
			out[k] = v
		}
	}

	return out
}
