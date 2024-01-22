package common

func MatchValueInList(val int, candidates []int, threshold int) int {
	for _, candidate := range candidates {
		if val > candidate - threshold && val < candidate + threshold {
			return candidate
		}
	}
	return -1
}
