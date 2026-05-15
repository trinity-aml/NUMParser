package utils

import (
	"NUMParser/db/models"
	"strconv"
	"strings"
)

func UniqueTorrList(arr []*models.TorrentDetails) []*models.TorrentDetails {
	inResult := make(map[string]struct{}, len(arr))
	result := make([]*models.TorrentDetails, 0, len(arr))
	for _, t := range arr {
		hash := ClearStr(t.Name + t.GetNames() + strconv.Itoa(t.Year))
		if _, ok := inResult[hash]; !ok {
			inResult[hash] = struct{}{}
			result = append(result, t)
		}
	}
	return result
}

func Distinct[T any](arr []T, getHash func(e T) string) []T {
	inResult := make(map[string]struct{}, len(arr))
	result := make([]T, 0, len(arr))
	for _, t := range arr {
		hash := getHash(t)
		if _, ok := inResult[hash]; !ok {
			inResult[hash] = struct{}{}
			result = append(result, t)
		}
	}
	return result
}

func ClearStr(str string) string {
	str = strings.ToLower(str)
	var b strings.Builder
	b.Grow(len(str))
	for _, r := range str {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'а' && r <= 'я') || r == 'ё' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
