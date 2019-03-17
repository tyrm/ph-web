package files

import (
	"errors"

	"../config"
	"github.com/juju/loggo"
	"github.com/minio/minio-go"
)

var logger *loggo.Logger
var mc *minio.Client
var mcInitialized = false

var bucket = ""

var (
	ErrNotInit = errors.New("files not initialized")
)

func init() {
	newLogger := loggo.GetLogger("files")
	logger = &newLogger
}

// InitClient attempts to initialize the Minio Client
func InitClient(config config.Config, force bool) {
	if mcInitialized && !force {
		return
	}

	logger.Infof("Initializing file store")

	// Initialize minio client object.
	useSSL := true
	minioClient, err := minio.New(config.FilesEndpoint, config.FilesKeyID, config.FilesAccessKey, useSSL)
	if err != nil {
		logger.Errorf("Problem initializing minio client %s", err.Error())
		return
	}
	bucket = config.FilesBucket

	// Create Bucket if exists
	exists, err := minioClient.BucketExists(bucket)
	if err == nil && exists {
		logger.Debugf("Bucket %s exists", bucket)
	} else if err == nil && !exists {
		err = minioClient.MakeBucket(bucket, "us-east-1")
		logger.Infof("Created bucket %s", bucket)
		if err != nil {
			logger.Errorf("Error creating bucket: %v", err)
			return
		}
	} else {
		logger.Errorf("Error checking bucket: %v", err)
		return
	}

	mc = minioClient
	mcInitialized = true
	logger.Infof("File store initialized")

}

// IsInit returns true if Minio client is initialized
func IsInit() (bool) {
	return mcInitialized
}