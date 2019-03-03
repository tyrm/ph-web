package files

import (
	"bytes"
	"log"

	"github.com/minio/minio-go"
)

func PutBytes(objectName string, data *[]byte) (n int64, err error) {
	reader := bytes.NewReader(*data)
	objectSize := int64(len(*data))

	// Upload the file
	n, err = mc.PutObject(bucket, objectName, reader, objectSize, minio.PutObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	return
}