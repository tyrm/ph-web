package registry

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"io"
	"strings"

	"github.com/juju/loggo"
	"github.com/lib/pq"
)

const sqlGetChildrenByParentID = `
SELECT r1.id, r1.parent_id, r1.key, r1.value, r1.secure, r1.created_at, r1.updated_at,
sum(case when r2.parent_id = r1.id then 1 else 0 end) as children
FROM registry r1 LEFT JOIN registry r2 ON r1.id = r2.parent_id
WHERE r1.parent_id = $1
GROUP BY r1.id;`

const sqlGetRegistryByKeyParentID = `
SELECT r1.id, r1.parent_id, r1.key, r1.value, r1.secure, r1.created_at, r1.updated_at,
sum(case when r2.parent_id = r1.id then 1 else 0 end) as children
FROM registry r1 LEFT JOIN registry r2 ON r1.id = r2.parent_id
WHERE r1.key = $1 AND r1.parent_id = $2
GROUP BY r1.id;`

const sqlGetRegistryRoot = `
SELECT r1.id, r1.key, r1.value, r1.secure, r1.created_at, r1.updated_at,
sum(case when r2.parent_id = r1.id then 1 else 0 end) as children
FROM registry r1 LEFT JOIN registry r2 ON r1.id = r2.parent_id
WHERE r1.key = '{ROOT}' AND r1.parent_id IS NULL
GROUP BY r1.id;`

var db *sql.DB
var logger *loggo.Logger

var reg_key []byte

type RegistryEntry struct {
	ID       int
	ParentID int
	Key      string
	Value    string
	Secure   bool

	ChildCount int

	CreatedAt pq.NullTime
	UpdatedAt pq.NullTime
}

func Close() {
	db.Close()

	return
}

func Init(connectionString string, password string) {
	newLogger := loggo.GetLogger("registry")
	logger = &newLogger

	logger.Debugf("Connecting to Database")
	dbClient, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Criticalf("Coud not connect to database: %s", err)
		panic(err)
	}
	db = dbClient

	db.SetMaxIdleConns(5)

	// calculate hash
	reg_key = []byte(createHash(password))
}

func GetRegistryEntry(path string) (reg *RegistryEntry, err error) {
	regCursor, newErr := getRegistryRoot()
	if newErr != nil {
		err = newErr
		return
	}

	if path == "/" {
		reg = &regCursor
		return
	}

	parts := SplitPath(path)
	logger.Tracef("Got parts: %v", parts)

	for _, key := range parts {
		regCursor, newErr = getRegistryByKey(key, regCursor.ID)
		if newErr != nil {
			err = newErr
			return
		}
	}
	reg = &regCursor

	return
}

func GetChildrenByID(id int) (reg []*RegistryEntry, err error) {
	var newRegList []*RegistryEntry

	rows, err := db.Query(sqlGetChildrenByParentID, id)
	if err != nil {
		logger.Tracef("GetChildrenByID(%d) (%v, %v)", id, nil, err)
		return
	}
	for rows.Next() {
		var id int
		var parentID int
		var key string
		var value sql.NullString
		var secure bool
		var createdAt pq.NullTime
		var updatedAt pq.NullTime
		var childCount int

		err = rows.Scan(&id, &parentID, &key, &value, &secure, &createdAt, &updatedAt, &childCount)
		if err != nil {
			logger.Tracef("GetChildrenByID(%d) (%v, %v)", id, nil, err)
			return
		}

		newReg := RegistryEntry{
			ID:         id,
			ParentID:   parentID,
			Key:        key,
			Secure:     secure,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
			ChildCount: childCount,
		}

		newRegList = append(newRegList, &newReg)
	}

	reg = newRegList
	logger.Tracef("GetUsersPage(%d) ([%d]User, %v)", id, len(reg), nil)

	return
}

func SplitPath(path string) ([]string) {
	newParts := strings.Split(path, "/")
	return newParts[1:]
}

// privates
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func decrypt(data []byte) []byte {
	block, err := aes.NewCipher(reg_key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func encrypt(data []byte) []byte {
	block, _ := aes.NewCipher(reg_key)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func getRegistryRoot() (reg RegistryEntry, err error) {
	var id int
	var key string
	var value sql.NullString
	var secure bool
	var createdAt pq.NullTime
	var updatedAt pq.NullTime
	var childCount int

	err = db.QueryRow(sqlGetRegistryRoot).Scan(&id, &key, &value, &secure, &createdAt, &updatedAt, &childCount)

	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	reg = RegistryEntry{
		ID:        id,
		Key:       key,
		Secure:    secure,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ChildCount: childCount,
	}

	if value.Valid {
		reg.Value = value.String
	}

	logger.Tracef("getRegistryRoot() (%v, %v)", reg, err)
	return
}

func getRegistryByKey(searchKey string, pID int) (reg RegistryEntry, err error) {
	var id int
	var parentID int
	var key string
	var value sql.NullString
	var secure bool
	var createdAt pq.NullTime
	var updatedAt pq.NullTime
	var childCount int

	err = db.QueryRow(sqlGetRegistryByKeyParentID, searchKey, pID).Scan(&id, &parentID, &key, &value, &secure, &createdAt, &updatedAt, &childCount)

	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	reg = RegistryEntry{
		ID:        id,
		ParentID:  parentID,
		Key:       key,
		Secure:    secure,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ChildCount: childCount,
	}

	if value.Valid {
		reg.Value = value.String
	}

	logger.Tracef("getRegistryByKey(%s) (%v, %v)", searchKey, reg, err)
	return
}
