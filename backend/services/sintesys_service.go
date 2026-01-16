package services

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/go-resty/resty/v2"
	"gorm.io/datatypes"
)

type Sintesys interface {
	SendCallback(npm, tahun_id string, ukt string) error
	ScanNewCallback()
}

type sintesys struct {
	AppUrl string
	Token  string
}

func NewSintesys() Sintesys {
	return &sintesys{AppUrl: os.Getenv("SINTESYS_CALLBACK_URL"), Token: os.Getenv("SINTESYS_TOKEN")}
}

func (s *sintesys) SendCallback(npm, tahun_id string, ukt string) error {
	// Gunakan Resty
	client := resty.New()

	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	values := url.Values{}
	values.Set("npm", npm)
	values.Set("tahun_id", tahun_id)
	if max_sks, isCapped := utils.MaxSKSFromUkt(ukt); isCapped {
		values.Set("max_sks", strconv.Itoa(max_sks))
	}

	formBody := make(map[string]string)
	for key, val := range values {
		formBody[key] = val[0]
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Bearer "+s.Token).
		SetFormData(formBody).
		Post(s.AppUrl)

	utils.Log.Info("Sintesys SendCallback", "npm : ", npm, " : ", resp)

	// Tidak menyimpan ke database - hanya consume data dari DBPNBP (read-only)

	if err != nil {
		utils.Log.Infof("Error on send callback %v", err.Error())
		return fmt.Errorf("gagal mengirim request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		utils.Log.Info(string(resp.Body()), " status code:", resp.StatusCode(), " url:", s.AppUrl)
		utils.Log.Info("Response Body:", string(resp.Body()))
		return errors.New("gagal membuat request ke sintesys")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("gagal parsing respons: %w", err)
	}
	return nil
}

func (s *sintesys) ScanNewCallback() {
	// Worker dihapus - tidak ada operasi write ke database
	// Semua data hanya read-only dari DBPNBP
}

func (s *sintesys) ProccessFromCallback(callback models.PaymentCallback) (bool, string, error) {
	utils.Log.Info("Sampe kesini")
	encodedString, err := s.FindDataEncoded(callback.Request)
	utils.Log.Infof("after FindDataEncoded %v", encodedString)
	if err != nil {
		return false, "", err
	}
	secret := utils.NewEpnbp().GenerateJWTSecret()
	claims, err := utils.CheckJwtBySecret(encodedString, []byte(secret), false)
	if err != nil {
		return false, "", err
	}
	invoiceId, err := s.ExtractInvoiceID(claims)
	if err != nil {
		return false, "", err
	}

	epnbpRepo := repositories.NewEpnbpRepository(database.DBPNBP)
	invoice, err := epnbpRepo.FindByInvoiceId(strconv.Itoa(invoiceId))
	if err != nil {
		return false, "", err
	}

	tagihanRepo := repositories.NewTagihanRepository(database.DBPNBP, database.DBPNBP)
	_, err = tagihanRepo.FindStudentBillByID(strconv.Itoa(int(invoice.InvoiceID)))

	if err != nil {
		return false, "", err
	}

	//mahasiswaRepo := repositories.NewMahasiswaRepository(database.DB)
	//mahasiswa, err := mahasiswaRepo.FindByMhswID(studentBill.StudentID)

	//s.SendCallback()

	return true, "Success", nil

}

func (s *sintesys) ExtractInvoiceID(body map[string]interface{}) (int, error) {
	// Ambil 'payment' dulu
	paymentRaw, ok := body["payment"]
	if !ok {
		return 0, errors.New("field 'payment' tidak ditemukan")
	}

	paymentMap, ok := paymentRaw.(map[string]interface{})
	if !ok {
		return 0, errors.New("'payment' bukan map[string]interface{}")
	}

	// Ambil 'invoice_id' dari payment
	invoiceIDRaw, ok := paymentMap["invoice_id"]
	if !ok {
		return 0, errors.New("field 'invoice_id' tidak ditemukan di dalam payment")
	}

	// Konversi ke float64 dulu, karena hasil JSON decoding untuk angka default-nya float64
	invoiceIDFloat, ok := invoiceIDRaw.(float64)
	if !ok {
		return 0, errors.New("'invoice_id' bukan float64")
	}

	return int(invoiceIDFloat), nil
}

func (s *sintesys) FindDataEncoded(request datatypes.JSON) (string, error) {
	var requestMap map[string]interface{}
	if err := json.Unmarshal(request, &requestMap); err != nil {
		utils.Log.Info("Gagal unmarshal callback.Request:", err)
		return "", fmt.Errorf("Gagal unmarshal callback.Request: %w", err)
	}

	utils.Log.Info("requestMap->Body :", requestMap["body"])

	bodyRaw, ok := requestMap["body"]
	if !ok {
		utils.Log.Info("Field 'body' tidak ditemukan")
		return "", fmt.Errorf("Field 'body' tidak ditemukan")
	}

	bodyMap, ok := bodyRaw.(map[string]interface{})
	if !ok {
		utils.Log.Info("Tipe 'body' bukan map[string]interface{}")
		return "", fmt.Errorf("Tipe 'body' bukan map[string]interface{}")
	}

	// --- Ambil field 'data' jika ada ---
	data, ok := bodyMap["data"]
	if ok {
		utils.Log.Info("DATA ditemukan:", data.(string))
		return data.(string), nil
	}

	log.Println("Field 'data' tidak ditemukan dalam body")
	return "", fmt.Errorf("Field 'data' tidak dalam body")

}
