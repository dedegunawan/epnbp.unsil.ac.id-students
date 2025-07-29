package utils

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
)

type Epnbp interface {
	CreateInvoice(payload map[string]interface{}) (map[string]interface{}, error)
	GenerateJWTSecret() string
	EncodePayloadToJWT(payload map[string]interface{}) (string, error)
}

type epnbp struct {
	AppUrl    string `json:"app_url"`
	AppId     string `json:"app_id"`
	SecretKey string `json:"secret"`
}

func NewEpnbp() Epnbp {
	return &epnbp{
		AppUrl:    os.Getenv("EPNBP_URL"),
		AppId:     os.Getenv("EPNBP_APP_ID"),
		SecretKey: os.Getenv("EPNBP_SECRET_KEY"),
	}
}

// Fungsi untuk generate MD5 dari app_id.secret_key
func (e *epnbp) GenerateJWTSecret() string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s.%s", e.AppId, e.SecretKey)))
	return hex.EncodeToString(hash[:])
}

// Fungsi untuk encode payload menjadi JWT
func (e *epnbp) EncodePayloadToJWT(payload map[string]interface{}) (string, error) {
	// Buat token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))

	// Signed string
	epnbpSecret := e.GenerateJWTSecret()
	return token.SignedString([]byte(epnbpSecret))
}

// Fungsi utama menggunakan Resty
func (e *epnbp) CreateInvoice(payload map[string]interface{}) (map[string]interface{}, error) {
	// Encode payload ke JWT string
	jwtString, err := e.EncodePayloadToJWT(payload)
	if err != nil {
		return nil, fmt.Errorf("gagal encode JWT: %w", err)
	}

	// Encode sebagai form url-encoded
	// Encode form-url
	formBody := "data=" + jwtString

	// Gunakan Resty
	client := resty.New()

	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetHeader("x-app-id", e.AppId).
		SetHeader("x-secret-key", e.SecretKey).
		SetBody(formBody).
		Post(e.AppUrl + "/api/invoices/create")
	if err != nil {
		Log.Info("Error creating invoice %v", err.Error())
		return nil, fmt.Errorf("gagal mengirim request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		Log.Info(string(resp.Body()), " status code:", resp.StatusCode(), " url:", e.AppUrl+"api/invoices/create")
		Log.Info("Response Body:", string(resp.Body()))
		return nil, errors.New("gagal membuat invoice EPNBP")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("gagal parsing respons: %w", err)
	}

	return result, nil
}
