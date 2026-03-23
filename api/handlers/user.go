package handlers

import (
	"strconv"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService *services.UserService
	log         *logger.Logger
}

func NewUserHandler(db *gorm.DB, log *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(db, log),
		log:         log,
	}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	users, err := h.userService.ListUsers()
	if err != nil {
		response.InternalError(c, "获取用户列表失败")
		return
	}

	response.Success(c, users)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, "获取用户失败")
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		if err == services.ErrUsernameAlreadyExist {
			response.BadRequest(c, "用户名已存在")
			return
		}
		response.InternalError(c, "创建用户失败")
		return
	}

	response.Created(c, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		if err == services.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		if err == services.ErrUsernameAlreadyExist {
			response.BadRequest(c, "用户名已存在")
			return
		}
		response.InternalError(c, "更新用户失败")
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	currentUser, ok := GetUserFromSession(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	if currentUser.Role != "admin" {
		response.Forbidden(c, "需要管理员权限")
		return
	}

	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}

	if currentUser.ID == userID {
		response.BadRequest(c, "不能删除当前登录用户")
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		if err == services.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, "删除用户失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *UserHandler) requireAdmin(c *gin.Context) bool {
	user, ok := GetUserFromSession(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return false
	}
	if user.Role != "admin" {
		response.Forbidden(c, "需要管理员权限")
		return false
	}
	return true
}

func parseUserIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return 0, false
	}
	return uint(id), true
}
