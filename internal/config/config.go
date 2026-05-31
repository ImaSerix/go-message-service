package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DB                 DBConfig
	WSHub              WSHubConfig
	IDGenerator        IDGeneratorConfig
	MaxMessagePageSize int
}

type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

func (c DBConfig) FormatDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

type WSHubConfig struct {
	ReadLimit  int64
	PongWait   time.Duration
	PingPerion time.Duration
	WriteWait  time.Duration
}

func loadWSHubConfig() WSHubConfig {
	readLimit, err := strconv.ParseInt(mustEnv("WS_READLIMIT"), 10, 64)
	if err != nil {
		log.Printf("invalid WS_READLIMIT, fallback to 512: %v", err)
		readLimit = 512
	}

	pongWait, err := time.ParseDuration(mustEnv("WS_PONGWAIT"))
	if err != nil {
		log.Printf("invalid WS_PONGWAIT, fallback to 60s: %v", err)
		pongWait = 60 * time.Second
	}

	pingPeriod, err := time.ParseDuration(mustEnv("WS_PINGPERIOD"))
	if err != nil {
		log.Printf("invalid WS_PINGPERIOD, fallback to 54s: %v", err)
		pingPeriod = 54 * time.Second
	}

	writeWait, err := time.ParseDuration(mustEnv("WS_WRITEWAIT"))
	if err != nil {
		log.Printf("invalid WS_WRITEWAIT, fallback to 10s: %v", err)
		pingPeriod = 10 * time.Second
	}

	return WSHubConfig{
		ReadLimit:  readLimit,
		PongWait:   pongWait,
		PingPerion: pingPeriod,
		WriteWait:  writeWait,
	}
}

type IDGeneratorConfig struct {
	NodeID int64
	Epoch  int64
}

func loadIDGeneratorConfig() IDGeneratorConfig {
	nodeID, err := strconv.ParseInt(mustEnv("IDGENERATOR_NODE_ID"), 10, 64)
	if err != nil {
		log.Printf("invalid IDGENERATOR_NODE_ID, fallback to 0: %v", err)
		nodeID = 0
	}

	epoch, err := strconv.ParseInt(mustEnv("IDGENERATOR_EPOCH"), 10, 64)
	if err != nil {
		log.Printf("invalid IDGENERATOR_EPOCH, fallback to 1700000000: %v", err)
		nodeID = 1700000000
	}

	return IDGeneratorConfig{
		NodeID: nodeID,
		Epoch:  epoch,
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing env %s", key)
	}
	return v
}

func Load() *Config {

	port, err := strconv.Atoi(mustEnv("DB_PORT"))
	if err != nil {
		log.Printf("invalid DB_PORT, fallback to 5432: %v", err)
		port = 5432
	}

	maxMessagePageSize, err := strconv.Atoi(mustEnv("MAX_MESSAGE_PAGE_SIZE"))
	if err != nil {
		log.Printf("invalid MAX_MESSAGE_PAGE_SIZE, fallback to 100: %v", err)
		maxMessagePageSize = 100
	}

	return &Config{
		DB: DBConfig{
			Host:     mustEnv("DB_HOST"),
			Port:     port,
			Name:     mustEnv("DB_NAME"),
			User:     mustEnv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  mustEnv("DB_SSLMODE"),
		},
		WSHub:              loadWSHubConfig(),
		IDGenerator:        loadIDGeneratorConfig(),
		MaxMessagePageSize: maxMessagePageSize,
	}
}
