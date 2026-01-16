package controllers

import (
	"encoding/json"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func PaymentCallbackHandler(c *gin.Context) {
	utils.Log.Info("PaymentCallbackHandler")
	// === 1. Ambil Header ===
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// === 2. Ambil Query Params ===
	queryParams := map[string]string{}
	for key, val := range c.Request.URL.Query() {
		if len(val) > 0 {
			queryParams[key] = val[0]
		}
	}

	// === 3. Ambil Body ===
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	var bodyData interface{}
	if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
		bodyData = string(bodyBytes) // fallback: raw body jika bukan JSON
	}

	// === 4. Response ke provider (customize sesuai kebutuhan) ===
	responseData := map[string]interface{}{
		"status":  "ok",
		"message": "callback received",
	}

	// Tidak menyimpan ke database - hanya consume data dari DBPNBP (read-only)

	// === 7. Kirim response ke provider ===
	c.JSON(http.StatusOK, responseData)
}
