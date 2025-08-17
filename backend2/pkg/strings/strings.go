package strings

import s "strings"

func GetEmailPrefix(email string) string {
	parts := s.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
