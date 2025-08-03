package utils

import "strconv"

func MaxSKSFromUkt(ukt string) (int, bool) {
	uktNumber, err := strconv.Atoi(ukt)
	if err != nil {
		return 0, false
	}
	if (uktNumber > 10 && uktNumber < 20) || (uktNumber > 1100 && uktNumber < 9999) {
		return 6, true
	}
	return 0, false

}
