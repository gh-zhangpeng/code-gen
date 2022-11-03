package utils

func Remove(strings []string, strs ...string) []string {
	out := append([]string(nil), strings...)
	for _, str := range strs {
		var n int
		for _, v := range out {
			if v != str {
				out[n] = v
				n++
			}
		}
		out = out[:n]
	}
	return out
}
