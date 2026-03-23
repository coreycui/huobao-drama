package handlers

import (
	"strconv"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StyleHandler struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewStyleHandler(db *gorm.DB, log *logger.Logger) *StyleHandler {
	return &StyleHandler{db: db, log: log}
}

func (h *StyleHandler) ListStyles(c *gin.Context) {
	var styles []models.ImageStyle
	if err := h.db.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Find(&styles).Error; err != nil {
		response.InternalError(c, "获取风格列表失败")
		return
	}
	response.Success(c, styles)
}

func (h *StyleHandler) ListAllStyles(c *gin.Context) {
	var styles []models.ImageStyle
	if err := h.db.Order("sort_order ASC, id ASC").Find(&styles).Error; err != nil {
		response.InternalError(c, "获取风格列表失败")
		return
	}
	response.Success(c, styles)
}

func (h *StyleHandler) CreateStyle(c *gin.Context) {
	var req struct {
		NameZH     string `json:"name_zh" binding:"required"`
		NameEN     string `json:"name_en" binding:"required"`
		StyleValue string `json:"style_value" binding:"required"`
		SortOrder  int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	style := models.ImageStyle{
		NameZH:     req.NameZH,
		NameEN:     req.NameEN,
		StyleValue: req.StyleValue,
		SortOrder:  req.SortOrder,
		IsActive:   true,
	}
	if err := h.db.Create(&style).Error; err != nil {
		response.InternalError(c, "创建风格失败")
		return
	}
	response.Created(c, style)
}

func (h *StyleHandler) UpdateStyle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var style models.ImageStyle
	if err := h.db.First(&style, id).Error; err != nil {
		response.NotFound(c, "风格不存在")
		return
	}

	var req struct {
		NameZH     string `json:"name_zh"`
		NameEN     string `json:"name_en"`
		StyleValue string `json:"style_value"`
		SortOrder  *int   `json:"sort_order"`
		IsActive   *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.NameZH != "" {
		updates["name_zh"] = req.NameZH
	}
	if req.NameEN != "" {
		updates["name_en"] = req.NameEN
	}
	if req.StyleValue != "" {
		updates["style_value"] = req.StyleValue
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.db.Model(&style).Updates(updates).Error; err != nil {
		response.InternalError(c, "更新失败")
		return
	}
	h.db.First(&style, id)
	response.Success(c, style)
}

func (h *StyleHandler) DeleteStyle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.db.Delete(&models.ImageStyle{}, id).Error; err != nil {
		response.InternalError(c, "删除失败")
		return
	}
	response.Success(c, nil)
}
