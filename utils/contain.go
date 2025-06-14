package utils

func Contains(i1 []int64, i2 int64) bool {
	for _, v := range i1 {
		if v == i2 {
			return true
		}
	}
	return false
}
