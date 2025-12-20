package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Env        string `mapstructure:"env"`
		KafkaTopic string `mapstructure:"kafka_topic"`
	} `mapstructure:"app"`

	FinAPI struct {
		HTTPHost   string `mapstructure:"http_host"`
		HTTPPort   int    `mapstructure:"http_port"`
		GRPCHost   string `mapstructure:"grpc_host"`
		GRPCPort   int    `mapstructure:"grpc_port"`
		GRPCTarget string `mapstructure:"grpc_target"`
	} `mapstructure:"fin_api"`

	FinAnalytics struct {
		HTTPHost string `mapstructure:"http_host"`
		HTTPPort int    `mapstructure:"http_port"`
	} `mapstructure:"fin_analytics"`

	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
}

type PostgresShardConfig struct {
	Name    string `mapstructure:"name"`
	ConnURL string `mapstructure:"conn_url"`
	Buckets int    `mapstructure:"buckets"`
}

type PostgresConfig struct {
	Shards []PostgresShardConfig `mapstructure:"shards"`
}

func (p PostgresConfig) GetTotalBuckets() int {
	total := 0
	for _, shard := range p.Shards {
		if shard.Buckets <= 0 {
			shard.Buckets = 1
		}
		total += shard.Buckets
	}
	return total
}

type RedisConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	DB   int    `mapstructure:"db"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	GroupID string   `mapstructure:"group_id"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvPrefix("fintrack")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetDefault("app.env", "development")
	v.SetDefault("app.kafka_topic", "user-transactions")
	v.SetDefault("fin_api.http_host", "0.0.0.0")
	v.SetDefault("fin_api.http_port", 8080)
	v.SetDefault("fin_api.grpc_host", "0.0.0.0")
	v.SetDefault("fin_api.grpc_port", 9090)
	v.SetDefault("fin_api.grpc_target", "fin-api:9090")
	v.SetDefault("fin_analytics.http_host", "0.0.0.0")
	v.SetDefault("fin_analytics.http_port", 8081)
	v.SetDefault("postgres.sslmode", "disable")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
