package files

import (
	"../registry"
	"github.com/juju/loggo"
	"github.com/minio/minio-go"
)

var logger *loggo.Logger
var mc *minio.Client
var mcInitialized = false

var bucket = ""

func init() {
	newLogger := loggo.GetLogger("files")
	logger = &newLogger
}

// InitClient attempts to initialize the Minio Client
func InitClient(force bool) {
	if mcInitialized && !force {
		return
	}

	logger.Infof("Initializing file store")
	var missingReg []string
	regEndpoint, err := registry.Get("/system/files/endpoint")
	if err != nil {
		if err == registry.ErrDoesNotExist {
			missingReg = append(missingReg, "endpoint")
		} else {
			logger.Errorf("Problem getting [endpoint]: %s", err.Error())
			return
		}
	}
	regKeyID, err := registry.Get("/system/files/key_id")
	if err != nil {
		if err == registry.ErrDoesNotExist {
			missingReg = append(missingReg, "key_id")
		} else {
			logger.Errorf("Problem getting [endpoint]: %s", err.Error())
			return
		}
	}
	regAccessKey, err := registry.Get("/system/files/access_key")
	if err != nil {
		if err == registry.ErrDoesNotExist {
			missingReg = append(missingReg, "access_key")
		} else {
			logger.Errorf("Problem getting [access_key]: %s", err.Error())
			return
		}
	}
	regBucket, err := registry.Get("/system/files/bucket")
	if err != nil {
		if err == registry.ErrDoesNotExist {
			missingReg = append(missingReg, "endpoint")
		} else {
			logger.Errorf("Problem getting [endpoint]: %s", err.Error())
			return
		}
	}

	if len(missingReg) > 0 {
		logger.Warningf("Could not init file system, missing registry items: %v", missingReg)
		return
	}

	endpoint, err := regEndpoint.GetValue()
	if err != nil {
		logger.Errorf("Problem getting [endpoint] value: %s", err.Error())
		return
	}
	secretBucket, err := regBucket.GetValue()
	if err != nil {
		logger.Errorf("Problem getting [bucket] value: %s", err.Error())
		return
	}
	accessKeyID, err := regKeyID.GetValue()
	if err != nil {
		logger.Errorf("Problem getting [key_id] value: %s", err.Error())
		return
	}
	secretAccessKey, err := regAccessKey.GetValue()
	if err != nil {
		logger.Errorf("Problem getting [access_key] value: %s", err.Error())
		return
	}

	// Initialize minio client object.
	useSSL := true
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		logger.Errorf("Problem initializing minio client %s", err.Error())
		return
	}
	bucket = secretBucket

	mc = minioClient
	mcInitialized = true
	logger.Infof("File store initialized")

}

// IsInit returns true if Minio client is initialized
func IsInit() (bool) {
	return mcInitialized
}