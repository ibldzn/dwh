package str

import "bytes"

func DetectDelimiter(sample []byte) rune {
	if len(sample) == 0 {
		return ','
	}

	candidates := []rune{',', ';', '|', '\t'}
	best := ','
	bestCount := -1

	sample = bytes.TrimPrefix(sample, []byte{0xEF, 0xBB, 0xBF})

	for _, candidate := range candidates {
		count := bytes.Count(sample, []byte(string(candidate)))
		if count > bestCount {
			bestCount = count
			best = candidate
		}
	}

	if bestCount <= 0 {
		return ','
	}
	return best
}
