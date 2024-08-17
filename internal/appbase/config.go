package appbase

import (
	"encoding/json"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Config represents common configuration for all applications.
type Config struct {
	// application config
	ServerAddress string `env:"SERVER_ADDRESS" env-default:"0.0.0.0:3000"`

	// application database
	ApplicationDatabaseName            string `env:"DATABASE_NAME" env-required:"true"`
	ApplicationDatabasePassword        string `env:"DATABASE_PASSWORD" env-required:"true"`
	ApplicationDatabasePort            string `env:"DATABASE_PORT" env-default:"5432"`
	ApplicationDatabaseUser            string `env:"DATABASE_USERNAME" env-required:"true"`
	ApplicationPrimaryDatabaseHost     string `env:"DATABASE_HOST" env-required:"true"`
	ApplicationReadReplicaDatabaseHost string `env:"DATABASE_HOST_RO" env-required:"true"`

	Env         string `env:"ENV"`
	LogLevel    string `env:"LOG_LEVEL"`
	HTTPTimeout int32  `env:"HTTP_TIMEOUT" env-default:"175"`

	// temporal connection
	TemporalHost                   string `env:"TEMPORAL_HOST"`
	TemporalNamespace              string `env:"TEMPORAL_NAMESPACE" env-default:"default"`
	TemporalTransfersTaskQueueName string `env:"TEMPORAL_TRANSFERS_TASK_QUEUE_NAME" env-default:"transfers"`

	//Redis
	RedisEndpoint           string `env:"REDIS_ENDPOINT" env-required:"true"`
	RedisPort               string `env:"REDIS_PORT" env-required:"true"`
	TransferMutexTTLSeconds int    `env:"TRANSFER_MUTEX_TTL_SECONDS" env-default:"300"`
}

func (c *Config) HTTPTimeoutDuration() time.Duration {
	return time.Duration(c.HTTPTimeout) * time.Second
}

func (c *Config) IsLogLevelDebug() bool {
	return c.LogLevel == zerolog.LevelDebugValue
}

func LoadConfig() (*Config, error) {
	c := new(Config)

	err := LoadConfiguration(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type goEnv struct {
	GoMod string `json:"GOMOD"`
}

func initLocal() (string, bool) {
	if !goExists() {
		return "", false
	}

	modRoot := getModuleRoot()
	envFilePath := fmt.Sprintf("%s/.env", modRoot)

	if envFileExists(envFilePath) {
		return envFilePath, true
	}

	return "", false
}

func envFileExists(envFilePath string) bool {
	_, err := os.Stat(envFilePath)

	return err == nil
}

func goExists() bool {
	_, err := exec.LookPath("go")

	return err == nil
}

func getModuleRoot() string {
	goEnvRaw, err := exec.Command("go", "env", "-json").Output()
	if err != nil {
		log.Fatal().Err(err).Msg("go env command failed")

		return ""
	}

	env := new(goEnv)

	err = json.Unmarshal(goEnvRaw, env)
	if err != nil {
		log.Fatal().Err(err).Msg("go mod unmarshalling failed")

		return ""
	}

	return strings.TrimSuffix(env.GoMod, "/go.mod")
}
func LoadConfiguration(cfg interface{}) error {
	envFilePath, fileExists := initLocal()

	var readCfgErr error

	if fileExists {
		readCfgErr = cleanenv.ReadConfig(envFilePath, cfg)
	} else {
		readCfgErr = cleanenv.ReadEnv(cfg)
	}

	if readCfgErr != nil {
		return readCfgErr
	}

	updateEnvErr := cleanenv.UpdateEnv(cfg)
	if updateEnvErr != nil {
		return updateEnvErr
	}

	return nil
}

type DBConfig struct {
	DatabaseName            string
	DatabasePassword        string
	DatabasePort            string
	DatabaseUser            string
	PrimaryDatabaseHost     string
	ReadReplicaDatabaseHost string
}
