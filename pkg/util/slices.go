package util

import (
	"strings"

	"github.com/samber/lo"
)

func SliceToMap(slice []string) map[string]string {
	return lo.SliceToMap(slice, func(s string) (string, string) {
		parts := strings.SplitN(s, "=", 2)
		return parts[0], parts[1]
	})
}
