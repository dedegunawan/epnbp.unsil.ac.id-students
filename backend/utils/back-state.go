package utils

import (
	"encoding/base64"
	"errors"
	"net/url"
)

func EncodeBackState(backTo string) string {
	stateParams := url.Values{}
	stateParams.Set("backTo", backTo)
	state := base64.URLEncoding.EncodeToString([]byte(stateParams.Encode()))
	return state
}

func DecodeBackState(encodedState string) (string, error) {
	decodedStateBytes, err := base64.URLEncoding.DecodeString(encodedState)
	if err != nil {
		return "", errors.New("Invalid state")
	}

	stateParams, err := url.ParseQuery(string(decodedStateBytes))
	if err != nil {
		return "", errors.New("Invalid state params")
	}

	backTo := stateParams.Get("backTo")

	return backTo, nil

}
