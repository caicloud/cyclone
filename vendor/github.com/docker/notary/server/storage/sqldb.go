package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// SQLStorage implements a versioned store using a relational database.
// See server/storage/models.go
type SQLStorage struct {
	gorm.DB
}

// NewSQLStorage is a convenience method to create a SQLStorage
func NewSQLStorage(dialect string, args ...interface{}) (*SQLStorage, error) {
	gormDB, err := gorm.Open(dialect, args...)
	if err != nil {
		return nil, err
	}
	return &SQLStorage{
		DB: gormDB,
	}, nil
}

// translateOldVersionError captures DB errors, and attempts to translate
// duplicate entry - currently only supports MySQL and Sqlite3
func translateOldVersionError(err error) error {
	switch err := err.(type) {
	case *mysql.MySQLError:
		// https://dev.mysql.com/doc/refman/5.5/en/error-messages-server.html
		// 1022 = Can't write; duplicate key in table '%s'
		// 1062 = Duplicate entry '%s' for key %d
		if err.Number == 1022 || err.Number == 1062 {
			return ErrOldVersion{}
		}
	}
	return err
}

// UpdateCurrent updates a single TUF.
func (db *SQLStorage) UpdateCurrent(gun string, update MetaUpdate) error {
	// ensure we're not inserting an immediately old version - can't use the
	// struct, because that only works with non-zero values, and Version
	// can be 0.
	exists := db.Where("gun = ? and role = ? and version >= ?",
		gun, update.Role, update.Version).First(&TUFFile{})

	if !exists.RecordNotFound() {
		return ErrOldVersion{}
	}
	checksum := sha256.Sum256(update.Data)
	return translateOldVersionError(db.Create(&TUFFile{
		Gun:     gun,
		Role:    update.Role,
		Version: update.Version,
		Sha256:  hex.EncodeToString(checksum[:]),
		Data:    update.Data,
	}).Error)
}

// UpdateMany atomically updates many TUF records in a single transaction
func (db *SQLStorage) UpdateMany(gun string, updates []MetaUpdate) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	rollback := func(err error) error {
		if rxErr := tx.Rollback().Error; rxErr != nil {
			logrus.Error("Failed on Tx rollback with error: ", rxErr.Error())
			return rxErr
		}
		return err
	}

	var (
		query *gorm.DB
		added = make(map[uint]bool)
	)
	for _, update := range updates {
		// This looks like the same logic as UpdateCurrent, but if we just
		// called, version ordering in the updates list must be enforced
		// (you cannot insert the version 2 before version 1).  And we do
		// not care about monotonic ordering in the updates.
		query = db.Where("gun = ? and role = ? and version >= ?",
			gun, update.Role, update.Version).First(&TUFFile{})

		if !query.RecordNotFound() {
			return rollback(ErrOldVersion{})
		}

		var row TUFFile
		checksum := sha256.Sum256(update.Data)
		hexChecksum := hex.EncodeToString(checksum[:])
		query = tx.Where(map[string]interface{}{
			"gun":     gun,
			"role":    update.Role,
			"version": update.Version,
		}).Attrs("data", update.Data).Attrs("sha256", hexChecksum).FirstOrCreate(&row)

		if query.Error != nil {
			return rollback(translateOldVersionError(query.Error))
		}
		// it's previously been added, which means it's a duplicate entry
		// in the same transaction
		if _, ok := added[row.ID]; ok {
			return rollback(ErrOldVersion{})
		}
		added[row.ID] = true
	}
	return tx.Commit().Error
}

// GetCurrent gets a specific TUF record
func (db *SQLStorage) GetCurrent(gun, tufRole string) (*time.Time, []byte, error) {
	var row TUFFile
	q := db.Select("updated_at, data").Where(
		&TUFFile{Gun: gun, Role: tufRole}).Order("version desc").Limit(1).First(&row)
	if err := isReadErr(q, row); err != nil {
		return nil, nil, err
	}
	return &(row.UpdatedAt), row.Data, nil
}

// GetChecksum gets a specific TUF record by its hex checksum
func (db *SQLStorage) GetChecksum(gun, tufRole, checksum string) (*time.Time, []byte, error) {
	var row TUFFile
	q := db.Select("created_at, data").Where(
		&TUFFile{
			Gun:    gun,
			Role:   tufRole,
			Sha256: checksum,
		},
	).First(&row)
	if err := isReadErr(q, row); err != nil {
		return nil, nil, err
	}
	return &(row.CreatedAt), row.Data, nil
}

func isReadErr(q *gorm.DB, row TUFFile) error {
	if q.RecordNotFound() {
		return ErrNotFound{}
	} else if q.Error != nil {
		return q.Error
	}
	return nil
}

// Delete deletes all the records for a specific GUN - we have to do a hard delete using Unscoped
// otherwise we can't insert for that GUN again
func (db *SQLStorage) Delete(gun string) error {
	return db.Unscoped().Where(&TUFFile{Gun: gun}).Delete(TUFFile{}).Error
}

// CheckHealth asserts that the tuf_files table is present
func (db *SQLStorage) CheckHealth() error {
	tableOk := db.HasTable(&TUFFile{})
	if db.Error != nil {
		return db.Error
	}
	if !tableOk {
		return fmt.Errorf(
			"Cannot access table: %s", TUFFile{}.TableName())
	}
	return nil
}
