package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const sessionCookieName = "session_id"

type sessionInfo struct {
	UserID    uint
	ExpiresAt time.Time
}

var (
	authUserService *services.UserService
	sessionStore    sync.Map
)

type AuthHandler struct {
	log *logger.Logger
}

func InitAuthService(db *gorm.DB, log *logger.Logger) {
	authUserService = services.NewUserService(db, log)
}

func NewAuthHandler(log *logger.Logger) *AuthHandler {
	return &AuthHandler{log: log}
}

func GetUserFromSession(c *gin.Context) (*models.User, bool) {
	if authUserService == nil {
		return nil, false
	}

	sessionID, err := c.Cookie(sessionCookieName)
	if err != nil || sessionID == "" {
		return nil, false
	}

	value, ok := sessionStore.Load(sessionID)
	if !ok {
		return nil, false
	}

	info, ok := value.(sessionInfo)
	if !ok || time.Now().After(info.ExpiresAt) {
		sessionStore.Delete(sessionID)
		return nil, false
	}

	user, err := authUserService.GetUserByID(info.UserID)
	if err != nil || !user.IsActive {
		sessionStore.Delete(sessionID)
		return nil, false
	}

	return user, true
}

func (h *AuthHandler) Login(c *gin.Context) {
	if authUserService == nil {
		response.InternalError(c, "认证服务未初始化")
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := authUserService.Authenticate(req.Username, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials || err == services.ErrUserDisabled {
			response.Unauthorized(c, "用户名或密码错误")
			return
		}
		response.InternalError(c, "登录失败")
		return
	}

	sessionID := uuid.NewString()
	expireAt := time.Now().Add(7 * 24 * time.Hour)
	sessionStore.Store(sessionID, sessionInfo{
		UserID:    user.ID,
		ExpiresAt: expireAt,
	})

	setSessionCookie(c, sessionID, int((7 * 24 * time.Hour).Seconds()))
	response.Success(c, gin.H{
		"user": user,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionCookieName)
	if err == nil && sessionID != "" {
		sessionStore.Delete(sessionID)
	}
	clearSessionCookie(c)
	response.Success(c, gin.H{"message": "退出成功"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := GetUserFromSession(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	response.Success(c, user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	if authUserService == nil {
		response.InternalError(c, "认证服务未初始化")
		return
	}

	user, ok := GetUserFromSession(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required,min=6,max=72"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := authUserService.ChangePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
		if err == services.ErrInvalidCredentials {
			response.BadRequest(c, "原密码错误")
			return
		}
		response.InternalError(c, "修改密码失败")
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}

func setSessionCookie(c *gin.Context, sessionID string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, sessionID, maxAge, "/", "", c.Request.TLS != nil, true)
}

func clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", c.Request.TLS != nil, true)
}
