package account

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Define salt size
const saltSize = 16

// Repository encapsulates the logic to access accounts from the data source.
type Repository interface {
	// Get returns the account with the specified username and password.
	Get(ctx context.Context, username, password string) (entity.Account, error)
	// GetByUsername returns the account with the specified username.
	GetByUsername(ctx context.Context, username string) (entity.Account, error)
	// Count returns the number of accounts.
	Count(ctx context.Context) (int, error)
	// Create saves a new account in the storage.
	Create(ctx context.Context, account entity.Account) error
	// Update updates the account with given ID in the storage.
	Update(ctx context.Context, account entity.Account) error
	// SetPassword updates the password for the given account id.
	SetPassword(ctx context.Context, id int, password string) error
	// Delete removes the account with given ID from the storage.
	Delete(ctx context.Context, id int) error
	// List returns a list of all entity.Account
	List(ctx context.Context) ([]entity.Account, error)
}

// repository persists accounts in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new account repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the account with the specified ID from the database.
func (r repository) Get(ctx context.Context, userName, password string) (entity.Account, error) {
	a, err := r.GetByUsername(ctx, userName)
	if err != nil {
		return a, err
	}

	if !r.isValidPassword(a, password) {
		return entity.Account{}, errors.New("invalid password")
	}

	return a, err
}

// GetByUsername reads the account with the specified username from the database.
func (r repository) GetByUsername(ctx context.Context, userName string) (entity.Account, error) {
	var accounts []entity.Account

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(&dbx.HashExp{"user_name": userName}).
		All(&accounts)

	if len(accounts) == 0 {
		return entity.Account{}, errors.New("sql: no rows in result set")
	}

	return accounts[0], err
}

// Create saves a new account record in the database.
// It returns the ID of the newly inserted account record.
func (r repository) Create(ctx context.Context, account entity.Account) error {
	// Generate the hashed password.
	salt, err := r.generateSafeRandomSalt(saltSize)
	if err != nil {
		return err
	}
	account.Salt = string(salt)
	account.HashedPassword = r.hashPassword(account.Password, []byte(account.Salt))

	return r.db.With(ctx).Model(&account).Insert()
}

// Update saves the changes to an account in the database.
func (r repository) Update(ctx context.Context, account entity.Account) error {
	if len(account.Password) == 0 {
		return errors.New("Invalid password")
	}

	// Generate the hashed password.
	salt, err := r.generateSafeRandomSalt(saltSize)
	if err != nil {
		return err
	}
	account.Salt = string(salt)
	account.HashedPassword = r.hashPassword(account.Password, []byte(account.Salt))
	account.UpdatedAt = time.Now()

	return r.db.With(ctx).Model(&account).Update()
}

// Delete deletes an account with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id int) error {
	account, err := r.getByID(ctx, id)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&account).Delete()
}

// Count returns the number of the account records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").From("account").Row(&count)
	return count, err
}

func (r repository) getByID(ctx context.Context, id int) (entity.Account, error) {
	var account entity.Account
	err := r.db.With(ctx).Select().Model(id, &account)
	return account, err
}

func (r repository) isValidPassword(a entity.Account, password string) bool {
	hp := r.hashPassword(password, []byte(a.Salt))

	return a.HashedPassword == hp
}

// SetPassword updates the password for the given account id.
func (r repository) SetPassword(ctx context.Context, id int, password string) error {
	salt, err := r.generateSafeRandomSalt(saltSize)
	if err != nil {
		return err
	}

	hashedPassword := r.hashPassword(password, salt)

	sql := "UPDATE account SET hashed_password='%s', salt='%s', requires_password_change=0, updated_at=DATE('now') WHERE id=%d"
	query := fmt.Sprintf(sql, hashedPassword, string(salt), id)
	_, err = r.db.DB().NewQuery(query).Execute()
	return err
}

// Combine password and salt then hash them using the SHA-512
// hashing algorithm and then return the hashed password
// as a hex string
func (r repository) hashPassword(password string, salt []byte) string {
	// Convert password string to byte slice
	var passwordBytes = []byte(password)

	// Create sha-512 hasher
	var sha512Hasher = sha512.New()

	// Append salt to password
	passwordBytes = append(passwordBytes, salt...)

	// Write password bytes to the hasher
	sha512Hasher.Write(passwordBytes)

	// Get the SHA-512 hashed password
	var hashedPasswordBytes = sha512Hasher.Sum(nil)

	// Convert the hashed password to a hex string
	var hashedPasswordHex = hex.EncodeToString(hashedPasswordBytes)

	return hashedPasswordHex
}

// Generate 16 bytes randomly and securely using the
// Cryptographically secure pseudorandom number generator (CSPRNG)
// in the crypto.rand package. It will enforce the returned string
// to not contain any "'" string on it.
func (r repository) generateSafeRandomSalt(saltSize int) ([]byte, error) {
	var salt = make([]byte, saltSize)

	for i := 0; i < 10; i++ {
		if _, err := rand.Read(salt[:]); err != nil {
			continue
		}

		if !bytes.Contains(salt, []byte("'")) {
			return salt, nil
		}
	}

	return salt, errors.New("could not generate a random salt")
}

// List returns a list of all entity.Account
func (r repository) List(ctx context.Context) ([]entity.Account, error) {
	var accounts []entity.Account
	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		OrderBy("created_at DESC").
		All(&accounts)
	return accounts, err
}
