package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const CurrentDBVersion = 4

type AppConfig struct {
	Installed bool        `json:"installed"`
	MySQL     MySQLConfig `json:"mysql"`
	Redis     RedisConfig `json:"redis"`
	JWTSecret string      `json:"jwt_secret"`
	Port      int         `json:"port"`
	DBVersion int         `json:"db_version"`
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

var (
	Conf     *AppConfig
	DB       *gorm.DB
	RDB      *redis.Client
	confFile = "data/config.json"
	mu       sync.RWMutex
)

func LoadConfig() error {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(confFile)
	if err != nil {
		if os.IsNotExist(err) {
			Conf = &AppConfig{
				Installed: false,
				Port:      3132,
			}
			return nil
		}
		return err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	Conf = &cfg
	return nil
}

func SaveConfig(cfg *AppConfig) error {
	mu.Lock()
	defer mu.Unlock()

	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	Conf = cfg
	return os.WriteFile(confFile, data, 0644)
}

func IsInstalled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return Conf != nil && Conf.Installed
}

func InitDatabase() error {
	if Conf == nil {
		return fmt.Errorf("config not loaded")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		Conf.MySQL.User,
		Conf.MySQL.Password,
		Conf.MySQL.Host,
		Conf.MySQL.Port,
		Conf.MySQL.Database,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return nil
}

func InitRedis() error {
	if Conf == nil {
		return fmt.Errorf("config not loaded")
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", Conf.Redis.Host, Conf.Redis.Port),
		Password: Conf.Redis.Password,
		DB:       Conf.Redis.DB,
	})

	return nil
}
