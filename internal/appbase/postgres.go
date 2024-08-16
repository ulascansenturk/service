package appbase

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

const (
	defaultConnectionLifetime = 10 * time.Minute
	defaultConnectionIdleTime = 5 * time.Minute
	defaultMaxPoolSize        = 25
	defaultSSLMode            = "disable"
)

type Database struct {
	GormDB        *gorm.DB
	credentials   *Credentials
	connectionCfg *ConnectionConfig
	serviceName   string
	sslMode       string
}

func (db *Database) PostgresServiceName() string {
	return fmt.Sprintf("%s.postgres", db.serviceName)
}

type Credentials struct {
	Name            string
	Password        string
	PrimaryHost     string
	ReadReplicaHost string
	User            string
	Port            string
}

func (c *Credentials) PostgresConfig(primary bool, sslMode string) postgres.Config {
	var host = c.ReadReplicaHost

	if primary {
		host = c.PrimaryHost
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
		sslMode,
	)

	return postgres.Config{DSN: dsn}
}

type ConnectionConfig struct {
	MaxLifetime        time.Duration
	MaxIdleTime        time.Duration
	MaxOpenConnections int
	MaxIdleConnections int
}

func NewPostgres(options ...func(*Database)) (*Database, error) {
	database := new(Database)
	for _, o := range options {
		o(database)
	}

	err := database.Init()
	if err != nil {
		return nil, err
	}

	return database, nil
}

func WithServiceName(serviceName string) func(database *Database) {
	return func(database *Database) {
		database.serviceName = serviceName
	}
}

func WithCredentials(credentials *Credentials) func(database *Database) {
	return func(database *Database) {
		database.credentials = credentials
	}
}

func WithConnectionConfig(cfg *ConnectionConfig) func(database *Database) {
	return func(database *Database) {
		database.connectionCfg = cfg
	}
}

func WithSSLMode(mode string) func(database *Database) {
	return func(database *Database) {
		database.sslMode = mode
	}
}

func (db *Database) Init() error {
	db.setDefaults()

	primaryConfig := db.credentials.PostgresConfig(true, db.sslMode)
	readReplicaConfig := db.credentials.PostgresConfig(false, db.sslMode)

	gormDB, gormDBErr := gorm.Open(
		postgres.Open(primaryConfig.DSN),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if gormDBErr != nil {
		return gormDBErr
	}

	gormReplicaErr := gormDB.Use(
		dbresolver.Register(
			dbresolver.Config{
				Replicas: []gorm.Dialector{postgres.Open(readReplicaConfig.DSN)},
				Policy:   dbresolver.RandomPolicy{},
			},
		),
	)
	if gormReplicaErr != nil {
		return gormReplicaErr
	}

	sqlDB, sqlDBErr := gormDB.DB()
	if sqlDBErr != nil {
		return sqlDBErr
	}

	sqlDB.SetMaxOpenConns(db.connectionCfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(db.connectionCfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(db.connectionCfg.MaxLifetime)
	sqlDB.SetConnMaxIdleTime(db.connectionCfg.MaxIdleTime)

	db.GormDB = gormDB

	return nil
}

func (db *Database) setDefaults() {
	if db.sslMode == "" {
		db.sslMode = defaultSSLMode
	}

	if db.connectionCfg.MaxOpenConnections == 0 {
		db.connectionCfg.MaxOpenConnections = defaultMaxPoolSize
	}

	if db.connectionCfg.MaxIdleConnections == 0 {
		db.connectionCfg.MaxIdleConnections = defaultMaxPoolSize
	}

	if db.connectionCfg.MaxLifetime == 0 {
		db.connectionCfg.MaxLifetime = defaultConnectionLifetime
	}

	if db.connectionCfg.MaxIdleTime == 0 {
		db.connectionCfg.MaxIdleTime = defaultConnectionIdleTime
	}
}
