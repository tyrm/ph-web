package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	AESSecret string
	DBEngine string
}

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

	// Validation
	if len(missingEnv) > 0 {
		var msg string = fmt.Sprintf("Environment variables missing: %v", missingEnv)
		logger.Criticalf(msg)
		panic(fmt.Sprint(msg))
	}

	return
}
