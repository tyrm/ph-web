package registry

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/juju/loggo"
	"github.com/lib/pq"
	"github.com/patrickmn/go-cache"
)

var db *sql.DB
var logger *loggo.Logger

var cPathByID *cache.Cache
var cRegByID *cache.Cache

var reg_key []byte

var ErrDoesNotExist = errors.New("regisrty entry doesn't exist")

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

func (r *RegistryEntry) Delete() (err error) {
	_, err = db.Exec(sqlDeleteByID, r.ID)

	//cPathByID.Delete()
	return
}

func (r *RegistryEntry) GetPath() (path string, err error) {
	path, err = GetPathByID(r.ID)
	return
}

func (r *RegistryEntry) GetValue() (v string, err error) {
	if r.Secure {
		bDec, newErr := base64.StdEncoding.DecodeString(r.Value)
		if err != nil {
			err = newErr
		}
		sDec := decrypt(bDec)
		v = string(sDec)
	} else {
		v = r.Value
	}
	return
}

func (r *RegistryEntry) SetValue(newValue string) (err error) {
	newEncodedValue := ""
	if r.Secure {
		bEnc := encrypt([]byte(newValue))
		sEnc := base64.URLEncoding.EncodeToString(bEnc)
		newEncodedValue = sEnc
	} else {
		newEncodedValue = newValue
	}

	var value sql.NullString
	err = db.QueryRow(sqlUpdateValue, r.ID, newEncodedValue).Scan(&value)
	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	r.Value = value.String
	return
}

func (r *RegistryEntry) String() string {
	return fmt.Sprintf("RegistryItem[%d->%d %s]", r.ID, r.ParentID, r.Key)
}

// publics
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

	// init cache
	cPathByID = cache.New(5*time.Minute, 10*time.Minute)
	cRegByID = cache.New(5*time.Minute, 10*time.Minute)
}

func Get(path string) (reg *RegistryEntry, err error) {
	regCursor, newErr := getRegistryRoot()
	if newErr != nil {
		err = newErr
		return
	}

	if path == "/" {
		reg = regCursor
		return
	}

	parts := SplitPath(path)

	for _, key := range parts {
		regCursor, newErr = getRegistryByKey(key, regCursor.ID)
		if newErr != nil {
			if newErr == sql.ErrNoRows {
				err = ErrDoesNotExist
			} else {
				err = newErr
			}
			return
		}
	}
	reg = regCursor
	return
}

func GetByID(id int) (reg *RegistryEntry, err error) {
	regNew, newErr := getRegistry(id)
	if newErr != nil {
		err = newErr
		return
	}

	reg = regNew
	logger.Tracef("GetByID(%d) (%v, %v)", id, reg, err)
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

func GetPathByID(id int) (path string, err error) {
	idStr := strconv.Itoa(id)
	if p, found := cPathByID.Get(idStr); found {
		path = p.(string)
		logger.Tracef("GetPathByID(%d) (%s, %s) [HIT]", id, path, err)
		return
	}

	rows, err := db.Query(sqlGetPathByID, id)
	if err != nil {
		logger.Tracef("GetChildrenByID(%d) (%v, %v)", id, nil, err)
		return
	}

	newPath := ""
	for rows.Next() {
		var id int
		var parentID sql.NullInt64
		var key string

		err = rows.Scan(&id, &parentID, &key)
		if err != nil {
			logger.Tracef("GetChildrenByID(%d) (%v, %v)", id, nil, err)
			return
		}

		if key != "{ROOT}" {
			newPath = "/" + key + newPath
		}
	}
	if newPath == "" {
		newPath = "/"
	}

	cPathByID.Set(idStr, newPath, cache.DefaultExpiration)
	path = newPath
	logger.Tracef("GetPathByID(%d) (%s, %s) [MISS]", id, path, err)
	return
}

func New(newPid int, newKey string, newValue string, newSecure bool, uid int) (reg *RegistryEntry, err error) {
	newEncodedValue := ""
	if newSecure {
		bEnc := encrypt([]byte(newValue))
		sEnc := base64.URLEncoding.EncodeToString(bEnc)
		newEncodedValue = sEnc
	} else {
		newEncodedValue = newValue
	}

	var id int
	var parentID sql.NullInt64
	var key string
	var value sql.NullString
	var secure bool
	var createdAt pq.NullTime
	var updatedAt pq.NullTime

	newCreatedAt := time.Now()
	err = db.QueryRow(sqlCreateEntry, newPid, newKey, newEncodedValue, newSecure, newCreatedAt, newCreatedAt).
		Scan(&id, &parentID, &key, &value, &secure, &createdAt, &updatedAt)
	if err != nil {
		logger.Errorf(err.Error())
		logger.Tracef("New(%d, %s, ***, %v) (%v, %v)", newPid, newKey, newSecure, reg, err)
		return
	}

	if secure {
		go logChange(id, uid, LogAddedSecure, "", "")
	} else {
		go logChange(id, uid, LogAdded, "", newEncodedValue)
	}

	reg = &RegistryEntry{
		ID:        id,
		Key:       key,
		Secure:    secure,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if parentID.Valid {
		reg.ParentID = int(parentID.Int64)
	}

	if value.Valid {
		reg.Value = value.String
	}
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

func getRegistry(searchId int) (reg *RegistryEntry, err error) {
	var id int
	var parentID sql.NullInt64
	var key string
	var value sql.NullString
	var secure bool
	var createdAt pq.NullTime
	var updatedAt pq.NullTime
	var childCount int

	err = db.QueryRow(sqlGetRegistryByID, searchId).Scan(&id, &parentID, &key, &value, &secure, &createdAt, &updatedAt, &childCount)
	if err != nil {
		logger.Errorf(err.Error())
		logger.Tracef("getRegistry(%d) (%v, %v)", searchId, reg, err)
		return
	}

	reg = &RegistryEntry{
		ID:         id,
		Key:        key,
		Secure:     secure,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		ChildCount: childCount,
	}

	if parentID.Valid {
		reg.ParentID = int(parentID.Int64)
	}

	if value.Valid {
		reg.Value = value.String
	}

	logger.Tracef("getRegistry(%d) (%v, %v)", searchId, reg, err)
	return
}

func getRegistryRoot() (reg *RegistryEntry, err error) {
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

	reg = &RegistryEntry{
		ID:         id,
		Key:        key,
		Secure:     secure,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		ChildCount: childCount,
	}

	if value.Valid {
		reg.Value = value.String
	}

	logger.Tracef("getRegistryRoot() (%v, %v)", reg, err)
	return
}

func getRegistryByKey(searchKey string, pID int) (reg *RegistryEntry, err error) {
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

	reg = &RegistryEntry{
		ID:         id,
		ParentID:   parentID,
		Key:        key,
		Secure:     secure,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		ChildCount: childCount,
	}

	if value.Valid {
		reg.Value = value.String
	}

	logger.Tracef("getRegistryByKey(%s) (%v, %v)", searchKey, reg, err)
	return
}
