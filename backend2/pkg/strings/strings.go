package strings

import (
	"errors"
	"strconv"
	s "strings"
)

func GetEmailPrefix(email string) string {
	parts := s.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func GetUint64FromAny(dataAny any) (uint64, error) {
	data := uint64(0)
	var err error
	switch v := dataAny.(type) {
	case int:
		data = uint64(v)
	case int64:
		data = uint64(v)
	case float64:
		data = uint64(v)
	case uint:
		data = uint64(v)
	case uint64:
		data = v
	case string:
		parsed, errConv := strconv.ParseUint(v, 10, 64)
		if errConv != nil {
			err = errConv
		}
		data = parsed
	default:
		err = errors.New("unsupported user ID type")
	}

	return data, err

}
