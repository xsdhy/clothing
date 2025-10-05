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

// RepositoryFactory creates the appropriate repository implementation based on database type
type RepositoryFactory struct{}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory() *RepositoryFactory {
	return &RepositoryFactory{}
}

// Helper function to initialize a repository
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

// CreateRepository creates the appropriate repository based on configuration
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

// createMySQLRepository creates a MySQL repository
func (f *RepositoryFactory) createMySQLRepository(cfg *config.Config) (Repository, error) {
	dsn := cfg.DSNURL
	if dsn == "" {
		// Construct DSN from individual components
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPassword, cfg.DBAddr, cfg.DBPort, cfg.DBName)
	}

	db, err := f.openGormDB(mysql.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Auto-migrate schema
	if err := f.migrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return sql.NewGormRepository(db), nil
}

// createSQLiteRepository creates a SQLite repository
func (f *RepositoryFactory) createSQLiteRepository(cfg *config.Config) (Repository, error) {
	filePath := cfg.DBPath
	if filePath == "" {
		filePath = "datas/clothing.db" // Default SQLite database file
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

	// Auto-migrate schema
	if err := f.migrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return sql.NewGormRepository(db), nil
}

// createPostgresRepository creates a PostgreSQL repository
func (f *RepositoryFactory) createPostgresRepository(cfg *config.Config) (Repository, error) {
	dsn := cfg.DSNURL
	if dsn == "" {
		// Construct DSN from individual components
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.DBAddr, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	}

	db, err := f.openGormDB(postgres.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Auto-migrate schema
	if err := f.migrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return sql.NewGormRepository(db), nil
}

func (f *RepositoryFactory) openGormDB(dialector gorm.Dialector) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second * 5,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Configure GORM
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // Use singular table names
		},
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// migrateSchema migrates the database schema
func (f *RepositoryFactory) migrateSchema(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.DbUsageRecord{},
	)
}
