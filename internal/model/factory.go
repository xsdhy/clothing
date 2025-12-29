package model

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"clothing/internal/model/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	DBTypeMySQL    = "mysql"
	DBTypeSQLite   = "sqlite"
	DBTypePostgres = "postgres"
)

// RepositoryFactory 根据数据库类型创建对应的仓库实现
type RepositoryFactory struct{}

// NewRepositoryFactory 创建新的仓库工厂
func NewRepositoryFactory() *RepositoryFactory {
	return &RepositoryFactory{}
}

// InitRepository 初始化仓库的辅助函数
func InitRepository(cfg *config.Config) (Repository, error) {
	factory := NewRepositoryFactory()

	if cfg.DBType == "" {
		return nil, nil
	}

	repo, err := factory.CreateRepository(cfg)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// CreateRepository 根据配置创建对应的仓库实现
func (f *RepositoryFactory) CreateRepository(cfg *config.Config) (Repository, error) {
	switch cfg.DBType {
	case DBTypeMySQL:
		return f.createMySQLRepository(cfg)
	case DBTypeSQLite:
		return f.createSQLiteRepository(cfg)
	case DBTypePostgres:
		return f.createPostgresRepository(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}
}

// createMySQLRepository 创建 MySQL 仓库
func (f *RepositoryFactory) createMySQLRepository(cfg *config.Config) (Repository, error) {
	dsn := cfg.DSNURL
	if dsn == "" {
		// 从各个配置项构建 DSN
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPassword, cfg.DBAddr, cfg.DBPort, cfg.DBName)
	}

	db, err := f.openGormDB(mysql.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// 自动迁移数据库表结构
	if err := f.migrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return sql.NewGormRepository(db), nil
}

// createSQLiteRepository 创建 SQLite 仓库
func (f *RepositoryFactory) createSQLiteRepository(cfg *config.Config) (Repository, error) {
	filePath := cfg.DBPath
	if filePath == "" {
		filePath = "datas/clothing.db" // 默认 SQLite 数据库文件
	}

	// 1) 在使用前确保多级目录存在
	//    注：SQLite 会在连接时自动创建 .db 文件，但前提是目录已存在
	if dir := filepath.Dir(filePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
	}

	db, err := f.openGormDB(sqlite.Open(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// 自动迁移数据库表结构
	if err := f.migrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return sql.NewGormRepository(db), nil
}

// createPostgresRepository 创建 PostgreSQL 仓库
func (f *RepositoryFactory) createPostgresRepository(cfg *config.Config) (Repository, error) {
	dsn := cfg.DSNURL
	if dsn == "" {
		// 从各个配置项构建 DSN
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.DBAddr, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	}

	db, err := f.openGormDB(postgres.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// 自动迁移数据库表结构
	if err := f.migrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return sql.NewGormRepository(db), nil
}

func (f *RepositoryFactory) openGormDB(dialector gorm.Dialector) (*gorm.DB, error) {
	// 配置 GORM 日志
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second * 5,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 配置 GORM
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// migrateSchema 迁移数据库表结构
func (f *RepositoryFactory) migrateSchema(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.DbUser{},
		&entity.DbUsageRecord{},
		&entity.DbProvider{},
		&entity.DbModel{},
		&entity.DbTag{},
		&entity.DbUsageRecordTag{},
	)
}
