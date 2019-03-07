package files

import (
	"io/ioutil"
	"log"

	"github.com/minio/minio-go"
)

func GetBytes(objectName string) (data *[]byte, err error) {
	if !mcInitialized {
		err = ErrNotInit
		return
	}

	// Upload the file
	obj, err := mc.GetObject(bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(obj)
	data = &body
	return
}