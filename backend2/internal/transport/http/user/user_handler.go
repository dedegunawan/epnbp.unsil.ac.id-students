package user

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/usecase"
	"net/http"
	"strconv"

	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/response"
	"github.com/gin-gonic/gin"
)

type avatarReq struct {
	Avatar string `json:"avatar" binding:"required,url"`
}

type activeReq struct {
	IsActive bool `json:"is_active"`
}

type UserHandler struct {
	usecases usecase.Usecase
}

func NewUserHandler(uc usecase.Usecase) *UserHandler {
	return &UserHandler{usecases: uc}
}

func (h *UserHandler) FindById(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	u, err := h.usecases.UserUsecase.GetByID(id)
	if err != nil {
		response.Error(c, 404, "not found")
		return
	}
	response.OK(c, u)
}
func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.usecases.UserUsecase.List(page, size)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"items": items, "total": total, "page": page, "size": size})
}
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req avatarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.usecases.UserUsecase.UpdateAvatar(id, req.Avatar); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id, "avatar": req.Avatar})
}
func (h *UserHandler) UpdateActiveStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req activeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.usecases.UserUsecase.SetActive(id, req.IsActive); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id, "is_active": req.IsActive})
}

func (h *UserHandler) RegisterRoute(rg *gin.RouterGroup) {

	g := rg.Group("/users")

	g.GET("/:id", h.FindById)
	g.GET("", h.GetUsers)
	g.PUT("/:id/avatar", h.UpdateAvatar)
	g.PUT("/:id/active", h.UpdateActiveStatus)
}
