package redis


import (
	"context"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
	ctx context.Context
}

//Redis的配置
type Config struct {
	Addr string
	Password string
	DB int  //数据库编号
	PoolSize int
	MinIdleConns int 
	DialTimeout time.Duration
	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

//DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addr:	"localhost:6379",
		Password:	"",
		DB:	0,
		PoolSize:	10,
		MinIdleConns:	5,
		DialTimeout: 5*time.Second,
		ReadTimeout: 3*time.Second,
		WriteTimeout: 3*time.Second,
	}
}


func NewClient(cfg *Config) (*Client, error) {
    if cfg == nil {
        cfg = DefaultConfig()
    }

    rdb := redis.NewClient(&redis.Options{
        Addr:         cfg.Addr,
        Password:     cfg.Password,
        DB:           cfg.DB,
        PoolSize:     cfg.PoolSize,
        MinIdleConns: cfg.MinIdleConns,
        DialTimeout:  cfg.DialTimeout,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
    })

    ctx := context.Background()

    // 测试连接
    if err := rdb.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("Redis 连接失败: %w", err)
    }

    return &Client{
        rdb: rdb,
        ctx: ctx,
    }, nil
}

func (c *Client)Close() error{
	return c.rdb.Close()
}



//一些常见的方法的封装
// Set 设置键值对
func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
    return c.rdb.Set(c.ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(key string) (string, error) {
    return c.rdb.Get(c.ctx, key).Result()
}

// GetInt 获取整数值
func (c *Client) GetInt(key string) (int, error) {
    val, err := c.rdb.Get(c.ctx, key).Int()
    if err == redis.Nil {
        return 0, nil
    }
    return val, err
}

// Del 删除键
func (c *Client) Del(keys ...string) error {
    return c.rdb.Del(c.ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(key string) (bool, error) {
    n, err := c.rdb.Exists(c.ctx, key).Result()
    return n > 0, err
}

// Expire 设置过期时间
func (c *Client) Expire(key string, expiration time.Duration) error {
    return c.rdb.Expire(c.ctx, key, expiration).Err()
}

// Incr 自增
func (c *Client) Incr(key string) (int64, error) {
    return c.rdb.Incr(c.ctx, key).Result()
}

// HSet 设置哈希字段
func (c *Client) HSet(key string, values ...interface{}) error {
    return c.rdb.HSet(c.ctx, key, values...).Err()
}

// HGet 获取哈希字段
func (c *Client) HGet(key, field string) (string, error) {
    return c.rdb.HGet(c.ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (c *Client) HGetAll(key string) (map[string]string, error) {
    return c.rdb.HGetAll(c.ctx, key).Result()
}

// LPush 左侧推入列表
func (c *Client) LPush(key string, values ...interface{}) error {
    return c.rdb.LPush(c.ctx, key, values...).Err()
}

// RPop 右侧弹出列表
func (c *Client) RPop(key string) (string, error) {
    return c.rdb.RPop(c.ctx, key).Result()
}

// 获取底层客户端（用于更复杂的操作）
func (c *Client) GetClient() *redis.Client {
    return c.rdb
}
