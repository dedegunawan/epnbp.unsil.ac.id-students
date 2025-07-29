package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger() {
	Log = logrus.New()

	// Aktifkan informasi pemanggil
	Log.SetReportCaller(true)

	// Coba buka file logs/app.log
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Log.Out = file
	} else {
		Log.Out = os.Stdout
		Log.Warn("❗ Failed to log to file, using default stdout")
	}

	// Custom formatter untuk tampilkan lokasi file & baris
	Log.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			// Ambil nama file dan baris, potong path jika terlalu panjang
			filename := f.File
			if idx := strings.LastIndex(filename, "/"); idx != -1 {
				filename = filename[idx+1:]
			}
			return "", filename + ":" + funcLine(f)
		},
	})

	Log.SetLevel(logrus.InfoLevel)
}

// funcLine membuat string gabungan function + line number
func funcLine(f *runtime.Frame) string {
	return fmt.Sprintf("%d", f.Line)
}
