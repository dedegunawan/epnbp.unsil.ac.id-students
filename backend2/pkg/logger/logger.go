package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"runtime/debug"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(level, filePath string) *Logger {
	// pastikan folder ada
	_ = os.MkdirAll(filepath.Dir(filePath), 0755)

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// parse level
	lvl := new(zapcore.Level)
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		*lvl = zapcore.InfoLevel
	}

	// ⛳️ INI BAGIAN YANG SALAH KEMARIN — HARUS TERIMA 3 RETURN
	fileWS, closeFn, err := zap.Open(filePath) // (WriteSyncer, func(), error)
	if err != nil {
		panic(fmt.Sprintf("open log file: %v", err))
	}
	// optional: tutup file saat app exit → panggil di main: defer lg.Sync(); defer closeFn()
	// tidak bisa defer di sini; simpan closeFn kalau mau dipakai nanti.
	_ = closeFn // jika tidak dipakai, hilangkan baris ini dan panggil di main.

	consoleWS := zapcore.Lock(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), fileWS, *lvl),       // ke file (JSON)
		zapcore.NewCore(zapcore.NewConsoleEncoder(encCfg), consoleWS, *lvl), // ke console
	)

	z := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &Logger{z.Sugar()}
}

func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}

func (l *Logger) RecoverPanic() {
	if r := recover(); r != nil {
		l.Errorf("panic recovered: %v", r)
		l.Errorf("stack trace:\n%s", getStackTrace())
	}
}

// Di bagian paling bawah file:
func getStackTrace() string {
	return string(debug.Stack())
}
