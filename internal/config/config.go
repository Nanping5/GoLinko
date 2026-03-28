package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type masterConfig struct {
	AppName string `mapstructure:"app_name" json:"app_name"`
	Host    string `mapstructure:"host" json:"host"`
	Port    int    `mapstructure:"port" json:"port"`
}

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
	DbName   string `mapstructure:"db_name" json:"db_name"`
	// 读写分离
	WriteHost string   `mapstructure:"write_host" json:"write_host"`
	WritePort int      `mapstructure:"write_port" json:"write_port"`
	ReadHosts []string `mapstructure:"read_hosts" json:"read_hosts"`
	ReadPorts []int    `mapstructure:"read_ports" json:"read_ports"`
	// 主从同步延迟容忍时间，用于读后写一致性
	ReplicationLag int `mapstructure:"replication_lag" json:"replication_lag"`
}
type RedisConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}
type AuthCodeConfig struct {
	AccessKeyId     string `mapstructure:"access_key_id" json:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" json:"access_key_secret"`
	SignName        string `mapstructure:"sign_name" json:"sign_name"`
	TemplateCode    string `mapstructure:"template_code" json:"template_code"`
}
type LogConfig struct {
	LogPath string `mapstructure:"log_path" json:"log_path"`
}

type KafkaConfig struct {
	MessageMode string        `mapstructure:"message_mode" json:"message_mode"`
	HostPort    string        `mapstructure:"hostport" json:"hostport"`
	LoginTopic  string        `mapstructure:"login_topic" json:"login_topic"`
	LogoutTopic string        `mapstructure:"logout_topic" json:"logout_topic"`
	ChatTopic   string        `mapstructure:"chat_topic" json:"chat_topic"`
	Partition   int           `mapstructure:"partition" json:"partition"`
	TimeOut     time.Duration `mapstructure:"timeout" json:"timeout"`
}
type StaticSrcConfig struct {
	StaticAvatarPath string `mapstructure:"static_avatar_path" json:"static_avatar_path"`
	StaticFilePath   string `mapstructure:"static_file_path" json:"static_file_path"`
}
type Smtp struct {
	EmailAddr  string `mapstructure:"email_addr" json:"email_addr"`
	SmtpKey    string `mapstructure:"smtp_key" json:"smtp_key"`
	SmtpServer string `mapstructure:"smtp_server" json:"smtp_server"`
}

// MinioConfig
type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint" json:"endpoint"`
	AccessKey string `mapstructure:"access_key" json:"access_key"`
	SecretKey string `mapstructure:"secret_key" json:"secret_key"`
	Bucket    string `mapstructure:"bucket" json:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl" json:"use_ssl"`
}

// DistributedConfig 分布式
type DistributedConfig struct {
	Enabled    bool   `mapstructure:"enabled" json:"enabled"`
	InstanceID string `mapstructure:"instance_id" json:"instance_id"`
}

type Config struct {
	MainConfig      masterConfig      `mapstructure:"main_config" json:"main_config"`
	MysqlConfig     MysqlConfig       `mapstructure:"mysql_config" json:"mysql_config"`
	RedisConfig     RedisConfig       `mapstructure:"redis_config" json:"redis_config"`
	AuthCodeConfig  AuthCodeConfig    `mapstructure:"auth_code_config" json:"auth_code_config"`
	LogConfig       LogConfig         `mapstructure:"log_config" json:"log_config"`
	KafkaConfig     KafkaConfig       `mapstructure:"kafka_config" json:"kafka_config"`
	StaticSrcConfig StaticSrcConfig   `mapstructure:"static_src_config" json:"static_src_config"`
	Smtp            Smtp              `mapstructure:"smtp" json:"smtp"`
	MinioConfig     MinioConfig       `mapstructure:"minio_config" json:"minio_config"`
	Distributed     DistributedConfig `mapstructure:"distributed" json:"distributed"`
}

var config *Config

func LoadConfig() error {
	v := viper.New()
	v.SetConfigName("configs")
	v.SetConfigType("toml")

	v.AddConfigPath("./configs")
	v.AddConfigPath("../../configs")
	v.AddConfigPath("../../../configs")
	v.AddConfigPath(".")

	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	cfg := new(Config)
	err = v.Unmarshal(cfg)
	if err != nil {
		return err
	}
	config = cfg
	fmt.Printf("配置加载成功 - Port: %d, Host: %s\n", config.MainConfig.Port, config.MainConfig.Host)
	return nil
}
func GetConfig() *Config {
	if config == nil {
		config = new(Config)
		_ = LoadConfig()
	}
	return config
}
