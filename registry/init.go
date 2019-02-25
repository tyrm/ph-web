package registry

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
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

var regKey []byte

var (
	ErrDoesNotExist = errors.New("regisrty entry doesn't exist")
)

// Entry represents a registry entry in the database
type Entry struct {
	ID       int
	ParentID int
	Key      string
	Value    []byte
	Secure   bool

	ChildCount int

	CreatedAt pq.NullTime
	UpdatedAt pq.NullTime
}

// Delete the registry entry
func (r *Entry) Delete() (err error) {
	_, err = db.Exec(sqlDeleteByID, r.ID)

	//cPathByID.Delete()
	return
}

// GetPath of the registry entry
func (r *Entry) GetPath() (path string, err error) {
	path, err = GetPathByID(r.ID)
	return
}

// GetValue of the registry entry
func (r *Entry) GetValue() (v string, err error) {
	if r.Secure {
		sDec := decrypt(r.Value)
		v = string(sDec)
	} else {
		v = string(r.Value)
	}
	return
}

// SetValue of the registry entry
func (r *Entry) SetValue(newValue string) (err error) {
	var newEncodedValue []byte
	if r.Secure {
		bEnc := encrypt([]byte(newValue))
		newEncodedValue = bEnc
	} else {
		newEncodedValue = []byte(newValue)
	}

	var value []byte
	err = db.QueryRow(sqlUpdateValue, r.ID, newEncodedValue).Scan(&value)
	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	r.Value = value
	return
}

func (r *Entry) String() string {
	return fmt.Sprintf("RegistryItem[%d->%d %s]", r.ID, r.ParentID, r.Key)
}

// publics

// Close cleans up database connections
func Close() {
	db.Close()
	return
}

// Init connects registry to database
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
	regKey = []byte(createHash(password))

	// init cache
	cPathByID = cache.New(5*time.Minute, 10*time.Minute)
	cRegByID = cache.New(5*time.Minute, 10*time.Minute)
}

// Get registry entry
func Get(path string) (reg *Entry, err error) {
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
			err = newErr
			return
		}
	}
	reg = regCursor
	return
}

// GetByID returns registry entry by id
func GetByID(id int) (reg *Entry, err error) {
	regNew, newErr := getRegistry(id)
	if newErr != nil {
		err = newErr
		return
	}

	reg = regNew
	logger.Tracef("GetByID(%d) (%v, %v)", id, reg, err)
	return
}

// GetChildrenByID returns registry entries by parent id
func GetChildrenByID(id int) (reg []*Entry, err error) {
	var newRegList []*Entry

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

		newReg := Entry{
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

// GetPathByID returns the registry path
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

// New registry entry in the database
func New(newPid int, newKey string, newValue string, newSecure bool, uid int) (reg *Entry, err error) {
	var newEncodedValue []byte
	if newSecure {
		bEnc := encrypt([]byte(newValue))
		newEncodedValue = bEnc
	} else {
		newEncodedValue = []byte(newValue)
	}

	var id int
	var parentID sql.NullInt64
	var key string
	var value []byte
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

	reg = &Entry{
		ID:        id,
		Key:       key,
		Secure:    secure,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if parentID.Valid {
		reg.ParentID = int(parentID.Int64)
	}

	if len(value) > 0 {
		reg.Value = value
	}
	return
}

// SplitPath into it's parts
func SplitPath(path string) []string {
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
	block, err := aes.NewCipher(regKey)
	if err != nil {
		logger.Errorf("decrypt: Error getting new cypher: %s", err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Errorf("decrypt: Error getting new GCM: %s", err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.Errorf("decrypt: Error decrypting: %s", err.Error())
	}
	return plaintext
}

func encrypt(data []byte) []byte {
	block, _ := aes.NewCipher(regKey)
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

func getRegistry(searchID int) (reg *Entry, err error) {
	var id int
	var parentID sql.NullInt64
	var key string
	var value []byte
	var secure bool
	var createdAt pq.NullTime
	var updatedAt pq.NullTime
	var childCount int

	err = db.QueryRow(sqlGetRegistryByID, searchID).Scan(&id, &parentID, &key, &value, &secure, &createdAt, &updatedAt, &childCount)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		logger.Tracef("getRegistry(%d) (%v, %v)", searchID, reg, err)
		return
	}

	reg = &Entry{
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

	if len(value) > 0 {
		reg.Value = value
	}

	logger.Tracef("getRegistry(%d) (%v, %v)", searchID, reg, err)
	return
}

func getRegistryRoot() (reg *Entry, err error) {
	var id int
	var key string
	var value []byte
	var secure bool
	var createdAt pq.NullTime
	var updatedAt pq.NullTime
	var childCount int

	err = db.QueryRow(sqlGetRegistryRoot).Scan(&id, &key, &value, &secure, &createdAt, &updatedAt, &childCount)
	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	reg = &Entry{
		ID:         id,
		Key:        key,
		Secure:     secure,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		ChildCount: childCount,
	}

	if len(value) > 0 {
		reg.Value = value
	}

	return
}

func getRegistryByKey(searchKey string, pID int) (reg *Entry, err error) {
	var id int
	var parentID int
	var key string
	var value []byte
	var secure bool
	var createdAt pq.NullTime
	var updatedAt pq.NullTime
	var childCount int

	err = db.QueryRow(sqlGetRegistryByKeyParentID, searchKey, pID).Scan(&id, &parentID, &key, &value, &secure, &createdAt, &updatedAt, &childCount)

	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	reg = &Entry{
		ID:         id,
		ParentID:   parentID,
		Key:        key,
		Secure:     secure,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		ChildCount: childCount,
	}

	if len(value) > 0 {
		reg.Value = value
	}
	return
}
