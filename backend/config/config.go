package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Admin  AdminConfig  `mapstructure:"admin"`
	WeChat WeChatConfig `mapstructure:"wechat"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

func (m *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.DBName, m.Charset)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type AdminConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Nickname string `mapstructure:"nickname"`
}

type WeChatConfig struct {
	UserAppID                       string `mapstructure:"user_appid"`
	UserSecret                      string `mapstructure:"user_secret"`
	UserCodeEnvVersion              string `mapstructure:"user_code_env_version"`
	MerchantAppID                   string `mapstructure:"merchant_appid"`
	MerchantSecret                  string `mapstructure:"merchant_secret"`
	MchID                           string `mapstructure:"mch_id"`
	MchAPIKey                       string `mapstructure:"mch_api_key"`
	NotifyURL                       string `mapstructure:"notify_url"`
	RefundCertP12Path               string `mapstructure:"refund_cert_p12_path"`
	RefundSyncIntervalSeconds       int    `mapstructure:"refund_sync_interval_seconds"`
	RefundSubscribePage             string `mapstructure:"refund_subscribe_page"`
	RefundReviewSubscribeTemplateID string `mapstructure:"refund_review_subscribe_template_id"`
	RefundReviewSubscribeOrderNoKey string `mapstructure:"refund_review_subscribe_order_no_key"`
	RefundReviewSubscribeAmountKey  string `mapstructure:"refund_review_subscribe_amount_key"`
	RefundReviewSubscribeStatusKey  string `mapstructure:"refund_review_subscribe_status_key"`
	RefundReviewSubscribeRemarkKey  string `mapstructure:"refund_review_subscribe_remark_key"`
	RefundResultSubscribeTemplateID string `mapstructure:"refund_result_subscribe_template_id"`
	RefundResultSubscribeOrderNoKey string `mapstructure:"refund_result_subscribe_order_no_key"`
	RefundResultSubscribeAmountKey  string `mapstructure:"refund_result_subscribe_amount_key"`
	RefundResultSubscribeStatusKey  string `mapstructure:"refund_result_subscribe_status_key"`
	RefundResultSubscribeRemarkKey  string `mapstructure:"refund_result_subscribe_remark_key"`
	PayMode                         string `mapstructure:"pay_mode"`
	MockPayResult                   string `mapstructure:"mock_pay_result"`
}

var AppConfig *Config

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %w", err)
	}
	AppConfig = &cfg
	return &cfg, nil
}
