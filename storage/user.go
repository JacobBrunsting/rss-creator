package storage

import (
	"fmt"
	"log"
	"strings"

	"github.com/rss-creator/models"
)

type user interface {
	CreateUser(user *models.User) error
	GetUser(username string) (*models.User, error)
	UpdateUser(username string, user *models.User) error
	DeleteUser(username string) error
	UpdateTokenValidity(username string, valid bool) error
}

func (d *sqlDb) CreateUser(user *models.User) error {
	_, err := d.db.Exec(`
        INSERT INTO Users (username, password, email) VALUES (?, ?, ?)
    `, user.Username, user.Password, user.Email)
	if err != nil {
		log.Printf("error inserting user %v into the database\n %v", user, err)
	}
	return err
}

func (d *sqlDb) GetUser(username string) (*models.User, error) {
	rows, err := d.db.Query(`
        SELECT Users.username, Users.password, Users.email, Users.invalidatedtokens FROM Users
		WHERE Users.username = ?
    `, username)
	if err != nil {
		log.Printf("error reading user %v from database\n%v", username, err)
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		u := &models.User{}
		err := rows.Scan(&u.Username, &u.Password, &u.Email, &u.InvalidatedTokens)
		if err != nil {
			log.Printf("error parsing database rows\n%v", err)
			return nil, err
		}
		return u, nil
	}

	return nil, &NotFound{fmt.Sprintf("user %v", username)}
}

func (d *sqlDb) UpdateUser(username string, user *models.User) error {
	values := []string{}
	args := make([]interface{}, 0)

	if user.Password != "" {
		values = append(values, "password = ?")
		args = append(args, user.Password)
	}

	if user.Email != "" {
		values = append(values, "email = ?")
		args = append(args, user.Email)
	}

	if len(args) == 0 {
		return nil
	}

	resp, err := d.db.Exec(`
        UPDATE Users SET `+strings.Join(values, ",")+` WHERE username = ?
    `, append(args, username)...)
	if err != nil {
		log.Printf("error updating user %v into the database\n %v", username, err)
	}

	rows, err := resp.RowsAffected()
	if err != nil {
		log.Printf("error getting number of rows affected by update\n %v", err)
		return err
	}

	if rows == 0 {
		return &NotFound{fmt.Sprintf("user %v", username)}
	}

	return err
}

func (d *sqlDb) DeleteUser(username string) error {
	resp, err := d.db.Exec(`DELETE FROM Users WHERE username = ?`, username)

	rows, err := resp.RowsAffected()
	if err != nil {
		log.Printf("error getting number of rows affected by delete\n %v", err)
		return err
	}

	if rows == 0 {
		return &NotFound{fmt.Sprintf("user %v", username)}
	}

	return nil
}

func (d *sqlDb) UpdateTokenValidity(username string, valid bool) error {
	resp, err := d.db.Exec(`
        UPDATE Users SET invalidatedtokens = ? WHERE username = ?
    `, !valid, username)
	if err != nil {
		log.Printf("error updating user %v token validity\n %v", username, err)
		return err
	}

	rows, err := resp.RowsAffected()
	if err != nil {
		log.Printf("error getting number of rows affected by token validity update\n %v", err)
		return err
	}

	if rows == 0 {
		return &NotFound{fmt.Sprintf("user %v", username)}
	}

	return err
}
