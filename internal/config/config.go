package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type MainConfig struct {
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
type Config struct {
	MainConfig      MainConfig      `mapstructure:"main_config" json:"main_config"`
	MysqlConfig     MysqlConfig     `mapstructure:"mysql_config" json:"mysql_config"`
	RedisConfig     RedisConfig     `mapstructure:"redis_config" json:"redis_config"`
	AuthCodeConfig  AuthCodeConfig  `mapstructure:"auth_code_config" json:"auth_code_config"`
	LogConfig       LogConfig       `mapstructure:"log_config" json:"log_config"`
	KafkaConfig     KafkaConfig     `mapstructure:"kafka_config" json:"kafka_config"`
	StaticSrcConfig StaticSrcConfig `mapstructure:"static_src_config" json:"static_src_config"`
	Smtp            Smtp            `mapstructure:"smtp" json:"smtp"`
}

var config *Config

func LoadConfig() error {
	v := viper.New()
	v.SetConfigName("configs")
	v.SetConfigType("toml")

	// 这里增加多种路径查找，以支持单元测试在不同目录运行
	v.AddConfigPath("./configs")        // root 运行
	v.AddConfigPath("../../configs")    // internal/service/gorms/ 运行
	v.AddConfigPath("../../../configs") // 更深层 运行
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
	// 调试：打印配置
	fmt.Printf("✅ 配置加载成功 - Port: %d, Host: %s\n", config.MainConfig.Port, config.MainConfig.Host)
	return nil
}
func GetConfig() *Config {
	if config == nil {
		config = new(Config) //若为空则重新加载配置文件
		_ = LoadConfig()
	}
	return config
}
