package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGContact struct {
	ID          int
	PhoneNumber string
	FirstName   string
	LastName    sql.NullString
	UserID      sql.NullInt64
	Vcard       sql.NullString
	CreatedAt   time.Time
	LastSeen    time.Time
}

const sqlUpdateTGContactLastSeen = `
UPDATE tg_contacts
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGContact) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGContactLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGContact = `
INSERT INTO "public"."tg_contacts" (phone_number, first_name, last_name, user_id, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;`

// CreateTGPhotoSize creates a new instance of a telegram user in the database.
func CreateTGContact(phoneNumber string, firstName string, lastName sql.NullString, userID sql.NullInt64) (tga *TGContact,
	err error) {

	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGContact, phoneNumber, firstName, lastName, userID, createdAt, createdAt).
		Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newContact := &TGContact{
		ID:          newID,
		PhoneNumber: phoneNumber,
		FirstName:   firstName,
		LastName:    lastName,
		UserID:      userID,
		CreatedAt:   createdAt,
		LastSeen:    createdAt,
	}
	tga = newContact
	return
}

func CreateTGContactFromAPI(apiContact *tgbotapi.Contact) (*TGContact, error) {

	lastName := sql.NullString{Valid: false}
	if apiContact.LastName != "" {
		lastName = sql.NullString{
			String: apiContact.LastName,
			Valid:  true,
		}
	}

	userID := sql.NullInt64{Valid: false}
	if apiContact.UserID > 0 {
		userID = sql.NullInt64{
			Int64: int64(apiContact.UserID),
			Valid: true,
		}
	}

	return CreateTGContact(apiContact.PhoneNumber, apiContact.FirstName, lastName, userID)
}

const sqlReadTGContactByMeta = `
SELECT id, phone_number, first_name, last_name, user_id, created_at, last_seen
FROM tg_contacts
WHERE phone_number = $1 AND first_name = $2 AND last_name = $3 AND user_id = $4;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGContactByMeta(phoneNumber string, firstName string, lastName sql.NullString, userID sql.NullInt64) (tga *TGContact, err error) {
	var newID int
	var newPhoneNumber string
	var newFirstName string
	var newLastName sql.NullString
	var newUserID sql.NullInt64
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGContactByMeta, phoneNumber, firstName, lastName, userID).Scan(&newID, &newPhoneNumber,
		&newFirstName, &newLastName, &newUserID, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newContact := &TGContact{
		ID:          newID,
		PhoneNumber: newPhoneNumber,
		FirstName:   newFirstName,
		LastName:    newLastName,
		UserID:      newUserID,
		CreatedAt:   newCreatedAt,
		LastSeen:    newLastSeen,
	}
	tga = newContact
	return
}
