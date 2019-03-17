package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/juju/loggo"
	"gopkg.in/ini.v1"
)

var logger *loggo.Logger

// Config represents configuration variables collected from system environment
type Config struct {
	AESSecret string

	Debug bool

	DBEngine string

	FilesAccessKey string
	FilesBucket string
	FilesEndpoint string
	FilesKeyID string

	LoggerConfig string

	StatsdAddress string
	StatsdPrefix  string

	TGToken string
}

// CollectConfig collects configuration variables from system environment
func CollectConfig() (config Config) {

	// CONFIG_FILE
	configFile := "/etc/ph-web.ini"
	var envConfigFile = os.Getenv("CONFIG_FILE")
	if envConfigFile != "" {
		configFile = envConfigFile
	}

	cfg, err := ini.Load(configFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%v", cfg)

	var missingEnv []string


	// DEFAULT
	var envAESSecret = cfg.Section("").Key("aes_secret").String()
	if envAESSecret != "" {
		config.AESSecret = envAESSecret
	} else {
		missingEnv = append(missingEnv, "aes_secret")
	}

	var envDBEngine = cfg.Section("").Key("db_engine").String()
	if envDBEngine != "" {
		config.DBEngine = envDBEngine
	} else {
		missingEnv = append(missingEnv, "db_engine")
	}

	// chatbot
	var envTGToken = cfg.Section("chatbot").Key("tg_token").String()
	if envTGToken != "" {
		config.TGToken = envTGToken
	} else {
		missingEnv = append(missingEnv, "chatbot:tg_token")
	}

	// debug
	var envDebug = cfg.Section("debug").Key("web").String()
	if strings.ToUpper(envDebug) == "TRUE" {
		config.Debug = true
	} else {
		config.Debug = false
	}

	var envLoggerLevel = cfg.Section("debug").Key("log_level").String()
	if envLoggerLevel == "" {
		config.LoggerConfig = "<root>=INFO"
	} else {
		config.LoggerConfig = fmt.Sprintf("<root>=%s", strings.ToUpper(envLoggerLevel))
	}

	// files
	var envFilesAccessKey = cfg.Section("files").Key("access_key").String()
	if envFilesAccessKey != "" {
		config.FilesAccessKey = envFilesAccessKey
	} else {
		missingEnv = append(missingEnv, "files:access_key")
	}

	var envFilesBucket = cfg.Section("files").Key("bucket").String()
	if envFilesBucket != "" {
		config.FilesBucket = envFilesBucket
	} else {
		missingEnv = append(missingEnv, "files:access_key")
	}

	var envFilesEndpoint = cfg.Section("files").Key("endpoint").String()
	if envFilesEndpoint != "" {
		config.FilesEndpoint = envFilesEndpoint
	} else {
		missingEnv = append(missingEnv, "files:endpoint")
	}

	var envFilesKeyID = cfg.Section("files").Key("key_id").String()
	if envFilesKeyID != "" {
		config.FilesKeyID = envFilesKeyID
	} else {
		missingEnv = append(missingEnv, "files:key_id")
	}

	// statsd
	var envStatsdAddress = cfg.Section("statsd").Key("address").String()
	if envStatsdAddress != "" {
		config.StatsdAddress = envStatsdAddress
	} else {
		config.StatsdAddress = "127.0.0.1:8125"
	}

	var envStatsdPrefix = cfg.Section("statsd").Key("address").String()
	if envStatsdPrefix != "" {
		config.StatsdPrefix = envStatsdPrefix
	} else {
		config.StatsdPrefix = "ph-web"
	}

	// Validation
	if len(missingEnv) > 0 {
		var msg = fmt.Sprintf("Config parameters missing: %v", missingEnv)
		logger.Criticalf(msg)
		panic(fmt.Sprint(msg))
	}

	return
}

func init() {
	newLogger := loggo.GetLogger("config")
	logger = &newLogger
}
