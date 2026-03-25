//一共六个职责:1.建立连接，2.管理连接池，3.健康检查，4.优雅关闭，5.事务管理，6.连接配置（超时或者最大连接数）

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
	
	_ "github.com/lib/pq" // PostgreSQL 驱动（匿名导入）
)



//Config配置
type Config struct {
	Host string
	Port int
	User string
	Password string
	Database string
	SSLMode  string // disable, require, verify-ca, verify-full

	//连接池配置
	MaxOpenConns int //最大打开连接数 
	MaxIdleConns int //最大空闲连接数
	ConnMaxLifetime time.Duration //连接最大存活时间
	ConnMaxIdleTime time.Duration //连接最大空闲时间

	//超时配置
	ConnectTimeout int //连接超时
	QueryTimeout int //查询超时
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		Database:        "atlas_routex",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnectTimeout:  10,
		QueryTimeout:    30,
	}
}



//创建新的连接
func NewConnection(cfg *Config) (*sql.DB, error) {
	// 1. 构建 DSN（数据源名称）
	dsn := buildDSN(cfg)
	
	// 2. 打开数据库连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// 3. 配置连接池
	configureConnectionPool(db, cfg)
	
	// 4. 验证连接是否可用
	if err := ping(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	log.Printf("Database connected successfully: %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	
	return db, nil
}



// buildDSN 构建 PostgreSQL 连接字符串
func buildDSN(cfg *Config) string {
	// 格式: postgres://user:password@host:port/database?sslmode=disable&connect_timeout=10
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
		cfg.ConnectTimeout,
	)
}

// configureConnectionPool 配置连接池参数
func configureConnectionPool(db *sql.DB, cfg *Config) {
	// 设置最大打开连接数
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	
	// 设置最大空闲连接数
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	
	// 设置连接最大存活时间
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	
	// 设置连接最大空闲时间
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}
}

// ping 检查数据库连接是否可用
func ping(db *sql.DB) error {
	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return db.PingContext(ctx)
}

// ========== 带超时的查询辅助函数 ==========

// QueryContextWithTimeout 带超时的查询
func QueryContextWithTimeout(db *sql.DB, query string, timeout time.Duration, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return db.QueryContext(ctx, query, args...)
}

// QueryRowContextWithTimeout 带超时的单行查询
func QueryRowContextWithTimeout(db *sql.DB, query string, timeout time.Duration, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return db.QueryRowContext(ctx, query, args...)
}

// ExecContextWithTimeout 带超时的执行
func ExecContextWithTimeout(db *sql.DB, query string, timeout time.Duration, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return db.ExecContext(ctx, query, args...)
}

// ========== 事务管理 ==========

// Transaction 事务封装
type Transaction struct {
	tx *sql.Tx
}

// BeginTx 开始事务
func BeginTx(db *sql.DB, ctx context.Context, opts *sql.TxOptions) (*Transaction, error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	return &Transaction{tx: tx}, nil
}

// Commit 提交事务
func (t *Transaction) Commit() error {
	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Rollback 回滚事务
func (t *Transaction) Rollback() error {
	if err := t.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

// Exec 在事务中执行
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// Query 在事务中查询
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// QueryRow 在事务中查询单行
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// ========== 优雅关闭 ==========

// Close 关闭数据库连接
func Close(db *sql.DB) error {
	if db == nil {
		return nil
	}
	
	log.Println("Closing database connection...")
	
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	
	log.Println("Database connection closed")
	return nil
}

// ========== 健康检查 ==========

// HealthCheck 健康检查
func HealthCheck(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	return db.PingContext(ctx)
}

// GetStats 获取连接池统计信息
func GetStats(db *sql.DB) sql.DBStats {
	return db.Stats()
}

// ========== 从环境变量加载配置 ==========

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *Config {
	cfg := DefaultConfig()
	
	// 这里可以从环境变量读取
	// 例如：cfg.Host = os.Getenv("DB_HOST")
	
	return cfg
}
