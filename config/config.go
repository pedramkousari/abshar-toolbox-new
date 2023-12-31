package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

const (
	updateTimeOut   = time.Minute * 15
	rollbackTimeOut = time.Minute * 5
)

type Config struct {
	Server           Server `mapstructure:"server"`
	DockerComposeDir string `mapstructure:"docker-compose-directory"`
	TempDir          string
	startedAt        time.Time
	UpdateTimeOut    time.Duration
	RollbackTimeOut  time.Duration
	EncryptKey       string
}

var config Config

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("can not get user home directory %v", err))
	}

	viper.AddConfigPath(home)
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("update-toolbox")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Not Exists update-toolbox.yaml")
	}

	cnf := new(Config)
	if err := viper.Unmarshal(cnf); err != nil {
		panic(fmt.Errorf("Unmarshal config Failed %v", err))
	}

	config = *cnf
	config.UpdateTimeOut = updateTimeOut
	config.RollbackTimeOut = rollbackTimeOut

	currentDirectory, _ := os.Getwd()
	config.TempDir = currentDirectory + "/temp"
	config.EncryptKey = "e10adc3949ba59abbe56e057f20f883e"
}

func (cnf *Config) SetStartTime() {
	cnf.startedAt = time.Now()
}

func (cnf *Config) GetStartTime() int64 {
	return cnf.startedAt.Unix()
}

func GetCnf() Config {
	return config
}

func Get(path string) string {
	return viper.GetString(path)
}
