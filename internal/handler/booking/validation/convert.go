package validation

import "strconv"

func Convert(valueStr string, defaultValue int) int {
	if valueStr == "" {
		return defaultValue
	}

	v, _ := strconv.Atoi(valueStr)
	return v
}
