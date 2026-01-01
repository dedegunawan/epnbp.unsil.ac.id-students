package server

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Engine struct {
	*gin.Engine
}

func New(lg *logger.Logger) *Engine {
	g := gin.New()
	return &Engine{g}
}
