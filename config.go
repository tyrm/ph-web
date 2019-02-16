package main

import (
	"fmt"
	"os"
)

type Config struct {
	DBEngine string
	DBEngineFile string

	AESSecret string
	AESSecretFile string
}

func CollectConfig() (config Config) {
	var missingEnv []string

	// DB_ENGINE
	config.DBEngine = os.Getenv("DB_ENGINE")
	config.DBEngineFile = os.Getenv("DB_ENGINE_FILE")
	if config.DBEngine == "" && config.DBEngineFile == "" {
		missingEnv = append(missingEnv, "DB_ENGINE")
	}

	// Validation
	if len(missingEnv) > 0 {
		var msg string = fmt.Sprintf("Environment variables missing: %v", missingEnv)
		logger.Criticalf(msg)
		panic(fmt.Sprint(msg))
	}

	return
}
