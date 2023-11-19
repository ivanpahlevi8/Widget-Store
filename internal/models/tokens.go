package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"log"
	"time"
)

const (
	ScopeAuth = "authentication"
)

// cereat token model
/**
token model dibuat dengan menggunakan for mat dari JWT token dimana terdapat head, kemudian payload dan token itu sendiri
- Untuk head sendiri akan berisi data mengenai scope yang digunakan pada token ini yang dalam hal ini adalah Scope dan
  plain text yang merupoakan informasi mengenai encoding dari token
- Unutk payload akan berisi data mengenai user serpeti user id, expirey date dari auth user
- kemudian terdapat token yang dalam hal ini adalah hash
*/
type Token struct {
	PlainText string    `json:"plain"`
	UserId    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expirey   time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// create function to generate toke
func GenerateToke(userId int64, expiry time.Duration, scope string) (*Token, error) {
	// create token object
	/**
	token pertama kali dibuat dengan mengimput data berupa payload dari tojkenb
	*/
	myToken := Token{
		UserId:  userId,
		Expirey: time.Now().Add(expiry),
		Scope:   scope,
	}

	// create random bytes
	/**
	bytes digunakan untuk menyimpan nilai dari hasil encoding
	*/
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)

	// check for an error
	if err != nil {
		return nil, err
	}

	// create token plain from base32
	/**
	encoding dilakukan dengan menggunakan base32
	*/
	plain := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)

	// generate token
	/**
	selanjutnya token dibuat dengan menggunakan algoritma encoding sha256
	*/
	tokenGenerate := sha256.Sum256([]byte(plain))

	// set value to token
	myToken.PlainText = plain
	myToken.Hash = tokenGenerate[:]

	// return value
	return &myToken, nil
}

// create function to add tokens to database
func (db *DbModel) AddTokenToDatabase(token *Token, user User) (int, error) {
	// create context to set request timeout
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close cancle
	defer cancle()

	// delete token first
	err := db.DeleteTokenFromDatabase(user.ID)

	// check for an error
	if err != nil {
		log.Println("error when querying to insert token into database : ", err)
		return -1, err
	}

	// create query
	queryTxt := `
		INSERT INTO tokens (user_id, name, email, token_hash, created_at, updated_at, expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	// execute query
	res, err := db.DbConn.ExecContext(
		ctx,
		queryTxt,
		user.ID,
		user.LastName,
		user.Email,
		token.Hash,
		time.Now(),
		time.Now(),
		token.Expirey,
	)

	// check for an error
	if err != nil {
		log.Println("error when querying to insert token into database : ", err)
		return -1, err
	}

	// get id
	getId, err := res.LastInsertId()

	// check for an error
	if err != nil {
		log.Println("error when getting last inserted id : ", err)
		return -1, err
	}

	// if success
	return int(getId), nil
}

// create function to delete tokens from database
func (db *DbModel) DeleteTokenFromDatabase(userId int) error {
	// create context to set request timeout
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close cancle
	defer cancle()

	// create query
	queryTxt := `
		DELETE FROM tokens WHERE user_id=?
	`

	// execute token
	_, err := db.DbConn.ExecContext(
		ctx,
		queryTxt,
		userId,
	)

	// check for an error
	if err != nil {
		log.Println("error when deleting tokens with user id : ", err)
		return err
	}

	// if success
	return nil
}

// create funtion to get user by tokens
func (db *DbModel) GetUserForTokens(token string) (*User, error) {
	// create context to set request timeout
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close cancle
	defer cancle()

	// create user object to hold value from dataase
	var getUser User

	// create query
	queryTxt := `
		SELECT u.id, u.first_name, u.last_name, u.email, u.password, u.created_at, u.updated_at
		FROM users u
		INNER JOIN tokens t ON (t.user_id=u.id)
		WHERE t.token_hash=? AND t.expiry > ?
	`

	// encoding token to equal with hash
	token_hash := sha256.Sum256([]byte(token))

	// execute query
	query := db.DbConn.QueryRowContext(ctx, queryTxt, token_hash[:], time.Now())

	// check for an error
	err := query.Err()

	// check for an error
	if err != nil {
		log.Println("error when querying user to get user by token : ", err)
		return nil, err
	}

	// scan user
	err = query.Scan(
		&getUser.ID,
		&getUser.FirstName,
		&getUser.LastName,
		&getUser.Email,
		&getUser.Password,
		&getUser.CreatedAt,
		&getUser.UpdatedAt,
	)

	// check for an error
	if err != nil {
		log.Println("error when scanning user from database : ", err)
		return nil, err
	}

	// if success
	return &getUser, nil
}
