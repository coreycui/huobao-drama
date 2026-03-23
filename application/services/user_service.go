package services

import (
	"errors"
	"strings"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserDisabled         = errors.New("user disabled")
	ErrUsernameAlreadyExist = errors.New("username already exists")
)

type UserService struct {
	db  *gorm.DB
	log *logger.Logger
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
	IsActive *bool  `json:"is_active"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=100"`
	Password string `json:"password" binding:"omitempty,min=6,max=72"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
	IsActive *bool  `json:"is_active"`
}

func NewUserService(db *gorm.DB, log *logger.Logger) *UserService {
	return &UserService{db: db, log: log}
}

func (s *UserService) ListUsers() ([]models.User, error) {
	var users []models.User
	if err := s.db.Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", strings.TrimSpace(username)).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*models.User, error) {
	username := strings.TrimSpace(req.Username)
	if _, err := s.GetUserByUsername(username); err == nil {
		return nil, ErrUsernameAlreadyExist
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := &models.User{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		IsActive:     isActive,
	}

	if err := s.db.Create(user).Error; err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrUsernameAlreadyExist
		}
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(id uint, req *UpdateUserRequest) (*models.User, error) {
	user, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Username != "" {
		username := strings.TrimSpace(req.Username)
		if username != user.Username {
			existing, err := s.GetUserByUsername(username)
			if err == nil && existing.ID != user.ID {
				return nil, ErrUsernameAlreadyExist
			}
			if err != nil && !errors.Is(err, ErrUserNotFound) {
				return nil, err
			}
		}
		updates["username"] = username
	}
	if req.Password != "" {
		passwordHash, err := HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = passwordHash
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return user, nil
	}

	if err := s.db.Model(user).Updates(updates).Error; err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrUsernameAlreadyExist
		}
		return nil, err
	}

	return s.GetUserByID(id)
}

func (s *UserService) DeleteUser(id uint) error {
	result := s.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *UserService) Authenticate(username, password string) (*models.User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	newHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.db.Model(user).Update("password_hash", newHash).Error
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "unique") || strings.Contains(errMsg, "duplicate")
}
