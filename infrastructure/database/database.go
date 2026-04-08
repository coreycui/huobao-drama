package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func NewDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	if cfg.Type == "sqlite" {
		dbDir := filepath.Dir(dsn)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	gormConfig := &gorm.Config{
		Logger: NewCustomLogger(),
	}

	var db *gorm.DB
	var err error

	if cfg.Type == "sqlite" {
		// 使用 modernc.org/sqlite 纯 Go 驱动（无需 CGO）
		// 添加并发优化参数：WAL 模式、busy_timeout、cache
		dsnWithParams := dsn + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&cache=shared"
		db, err = gorm.Open(sqlite.Dialector{
			DriverName: "sqlite",
			DSN:        dsnWithParams,
		}, gormConfig)
	} else {
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// SQLite 连接池配置（modernc.org/sqlite 驱动内部已做串行化，此处设置合理缓冲）
	if cfg.Type == "sqlite" {
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetMaxOpenConns(10) // 允许少量并发读，避免轮询被写操作阻塞
	} else {
		sqlDB.SetMaxIdleConns(cfg.MaxIdle)
		sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	}
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// SeedDefaultStyles seeds default image styles if the table is empty
func SeedDefaultStyles(db *gorm.DB) error {
	var count int64
	db.Model(&models.ImageStyle{}).Count(&count)
	if count > 0 {
		return nil
	}

	defaults := []models.ImageStyle{
		{NameZH: "吉卜力", NameEN: "Ghibli", StyleValue: "ghibli", SortOrder: 1},
		{NameZH: "国漫", NameEN: "Chinese Animation", StyleValue: "guoman", SortOrder: 2},
		{NameZH: "末日废土", NameEN: "Post-Apocalyptic", StyleValue: "wasteland", SortOrder: 3},
		{NameZH: "怀旧", NameEN: "Nostalgic", StyleValue: "nostalgia", SortOrder: 4},
		{NameZH: "像素艺术", NameEN: "Pixel Art", StyleValue: "pixel", SortOrder: 5},
		{NameZH: "方块世界", NameEN: "Voxel World", StyleValue: "voxel", SortOrder: 6},
		{NameZH: "都市", NameEN: "Urban", StyleValue: "urban", SortOrder: 7},
		{NameZH: "国漫3D", NameEN: "Chinese 3D Animation", StyleValue: "guoman3d", SortOrder: 8},
		{NameZH: "Q版3D", NameEN: "Chibi 3D", StyleValue: "chibi3d", SortOrder: 9},
		{NameZH: "真人写实", NameEN: "Realistic", StyleValue: "realistic", SortOrder: 10},
	}
	return db.Create(&defaults).Error
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 核心模型
		&models.Drama{},
		&models.Episode{},
		&models.Character{},
		&models.Scene{},
		&models.Storyboard{},
		&models.FramePrompt{},
		&models.Prop{},

		// 生成相关
		&models.ImageGeneration{},
		&models.VideoGeneration{},
		&models.VideoMerge{},

		// AI配置
		&models.AIServiceConfig{},
		&models.AIServiceProvider{},

		// 资源管理
		&models.Asset{},
		&models.CharacterLibrary{},

		// 任务管理
		&models.AsyncTask{},

		// 风格管理
		&models.ImageStyle{},

		// 用户认证
		&models.User{},
	)
}

func SeedDatabase(db *gorm.DB) error {
	return seedDefaultStyles(db)
}

func seedDefaultStyles(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.ImageStyle{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaultStyles := []models.ImageStyle{
		{NameZH: "吉卜力", NameEN: "Ghibli", StyleValue: "ghibli", SortOrder: 1, IsActive: true},
		{NameZH: "国漫", NameEN: "Chinese Animation", StyleValue: "guoman", SortOrder: 2, IsActive: true},
		{NameZH: "末日废土", NameEN: "Post-Apocalyptic", StyleValue: "wasteland", SortOrder: 3, IsActive: true},
		{NameZH: "怀旧", NameEN: "Nostalgic", StyleValue: "nostalgia", SortOrder: 4, IsActive: true},
		{NameZH: "像素艺术", NameEN: "Pixel Art", StyleValue: "pixel", SortOrder: 5, IsActive: true},
		{NameZH: "方块世界", NameEN: "Voxel World", StyleValue: "voxel", SortOrder: 6, IsActive: true},
		{NameZH: "都市", NameEN: "Urban", StyleValue: "urban", SortOrder: 7, IsActive: true},
		{NameZH: "国漫3D", NameEN: "Chinese 3D Animation", StyleValue: "guoman3d", SortOrder: 8, IsActive: true},
		{NameZH: "Q版3D", NameEN: "Chibi 3D", StyleValue: "chibi3d", SortOrder: 9, IsActive: true},
		{NameZH: "真人写实", NameEN: "Realistic", StyleValue: "realistic", SortOrder: 10, IsActive: true},
	}

	return db.Create(&defaultStyles).Error
}

func SeedAdminUser(db *gorm.DB) error {
	var user models.User
	err := db.Where("username = ?", "admin").First(&user).Error
	if err == nil {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	adminUser := &models.User{
		Username:     "admin",
		PasswordHash: string(passwordHash),
		Role:         "admin",
		IsActive:     true,
	}

	return db.Create(adminUser).Error
}
