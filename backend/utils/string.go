package utils

import "strings"

func GetEmailPrefix(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
