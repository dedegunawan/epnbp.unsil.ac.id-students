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
	SearchByInvoiceID(invoiceId string) (map[string]interface{}, error)
	SearchByVirtualAccount(virtualAccount string) (map[string]interface{}, error)
	SearchByIdentifier(identifier string) (map[string]interface{}, error)
}

type epnbp struct {
	AppUrl    string `json:"app_url"`
	AppId     string `json:"app_id"`
	SecretKey string `json:"secret"`
}

func NewEpnbp() Epnbp {
	appId := os.Getenv("EPNBP_APP_ID")
	secretKey := os.Getenv("EPNBP_SECRET_KEY")
	appUrl := os.Getenv("EPNBP_URL")
	
	// Use provided credentials if env vars not set
	if appId == "" {
		appId = "4094557b-3fbd-4762-892c-60250b0fd0f4"
	}
	if secretKey == "" {
		secretKey = "wsbSpgEphBaVwyo6TisBgyz47EH4gWzj8Ft20DWe4Ef7WDnlzytSc2RnGc2px7Ha"
	}
	if appUrl == "" {
		appUrl = "https://epnbp.unsil.ac.id" // Default URL
	}
	
	Log.Infof("EPNBP Config: URL=%s, AppID=%s (first 10 chars)", appUrl, appId[:min(10, len(appId))])
	
	return &epnbp{
		AppUrl:    appUrl,
		AppId:     appId,
		SecretKey: secretKey,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
		Log.Infof("Error creating invoice %v", err.Error())
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

// Fungsi utama menggunakan Resty
func (e *epnbp) SearchByInvoiceID(invoiceId string) (map[string]interface{}, error) {

	// Gunakan Resty
	client := resty.New()

	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetHeader("x-app-id", e.AppId).
		SetHeader("x-secret-key", e.SecretKey).
		SetQueryParam("invoice_id", invoiceId).
		Get(e.AppUrl + "/api/virtual-accounts/search-invoice")
	if err != nil {
		Log.Infof("Error creating search invoice %v", err.Error())
		return nil, fmt.Errorf("gagal mengirim request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		Log.Info(string(resp.Body()), " status code:", resp.StatusCode(), " url:", e.AppUrl+"/api/virtual-accounts/search-invoice", "invoice_id", invoiceId)
		Log.Info("Response Body:", string(resp.Body()))
		return nil, errors.New("gagal membuat seach invoice EPNBP")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("gagal parsing respons: %w", err)
	}

	return result, nil
}

// Fungsi utama menggunakan Resty
func (e *epnbp) SearchByVirtualAccount(virtualAccount string) (map[string]interface{}, error) {

	// Gunakan Resty
	client := resty.New()

	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetHeader("x-app-id", e.AppId).
		SetHeader("x-secret-key", e.SecretKey).
		SetQueryParam("va", virtualAccount).
		Get(e.AppUrl + "/api/virtual-accounts/search-va")
	if err != nil {
		Log.Infof("Error creating search virtual account %v", err.Error())
		return nil, fmt.Errorf("gagal mengirim request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		Log.Info(string(resp.Body()), " status code:", resp.StatusCode(), " url:", e.AppUrl+"/api/virtual-accounts/search-va")
		Log.Info("Response Body:", string(resp.Body()))
		return nil, errors.New("gagal membuat seach va EPNBP")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("gagal parsing respons: %w", err)
	}

	return result, nil
}

// SearchByIdentifier mencari data pembayaran berdasarkan identifier (NPM/Student ID)
func (e *epnbp) SearchByIdentifier(identifier string) (map[string]interface{}, error) {
	// Gunakan Resty
	client := resty.New()

	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetHeader("x-app-id", e.AppId).
		SetHeader("x-secret-key", e.SecretKey).
		SetQueryParam("identifier", identifier).
		Get(e.AppUrl + "/api/virtual-accounts/search-identifier")
	if err != nil {
		Log.Infof("Error searching by identifier %v", err.Error())
		return nil, fmt.Errorf("gagal mengirim request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		Log.Info(string(resp.Body()), " status code:", resp.StatusCode(), " url:", e.AppUrl+"/api/virtual-accounts/search-identifier", "identifier", identifier)
		Log.Info("Response Body:", string(resp.Body()))
		return nil, errors.New("gagal search by identifier EPNBP")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("gagal parsing respons: %w", err)
	}

	return result, nil
}
