package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Config represents configuration variables collected from system environment
type Config struct {
	AESSecret     string
	Debug         bool
	DBEngine      string
	LoggerConfig  string
	StatsdAddress string
	StatsdPrefix  string
}

// CollectConfig collects configuration variables from system environment
func CollectConfig() (config Config) {
	var missingEnv []string

	// AES_SECRET
	config.AESSecret = os.Getenv("AES_SECRET")
	if config.AESSecret == "" {
		AESSecretFile := os.Getenv("AES_SECRET_FILE")
		if AESSecretFile == "" {
			AESSecretFile = "/run/secrets/ph_web_aes_secret"
		}

		data, err := ioutil.ReadFile(AESSecretFile)
		if err != nil {
			missingEnv = append(missingEnv, "AES_SECRET")
			missingEnv = append(missingEnv, "AES_SECRET_FILE")
		} else {
			config.AESSecret = string(data)
		}
	}

	// LOGGER_LEVEL
	var envDebug = os.Getenv("DEBUG")

	if strings.ToUpper(envDebug) == "TRUE" {
		config.Debug = true
	} else {
		config.Debug = false
	}

	// DB_ENGINE
	config.DBEngine = os.Getenv("DB_ENGINE")
	if config.DBEngine == "" {
		DBEngineFile := os.Getenv("DB_ENGINE_FILE")
		if DBEngineFile == "" {
			DBEngineFile = "/run/secrets/ph_web_db_engine"
		}

		data, err := ioutil.ReadFile(DBEngineFile)
		if err != nil {
			missingEnv = append(missingEnv, "DB_ENGINE")
			missingEnv = append(missingEnv, "DB_ENGINE_FILE")
		} else {
			config.DBEngine = string(data)
		}
	}

	// LOGGER_LEVEL
	var envLoggerLevel = os.Getenv("LOGGER_LEVEL")

	if envLoggerLevel == "" {
		config.LoggerConfig = "<root>=INFO"
	} else {
		config.LoggerConfig = fmt.Sprintf("<root>=%s", envLoggerLevel)
	}

	// STATSD_ADDR
	var envStatsdAddress = os.Getenv("STATSD_ADDR")

	if envStatsdAddress == "" {
		config.StatsdAddress = "127.0.0.1:8125"
	} else {
		config.StatsdAddress = envStatsdAddress
	}

	// STATSD_PREFIX
	var envStatsdPrefix = os.Getenv("STATSD_PREFIX")

	if envStatsdPrefix == "" {
		config.StatsdPrefix = "ph-web"
	} else {
		config.StatsdPrefix = envStatsdPrefix
	}

	// Validation
	if len(missingEnv) > 0 {
		var msg = fmt.Sprintf("Environment variables missing: %v", missingEnv)
		logger.Criticalf(msg)
		panic(fmt.Sprint(msg))
	}

	return
}
