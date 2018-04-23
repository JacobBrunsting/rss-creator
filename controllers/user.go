package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/rss-creator/models"
	"github.com/rss-creator/storage"
	"github.com/rss-creator/utils"
)

const (
	passMinLength = 8
)

type UserController interface {
	PostUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	GetUserExists(w http.ResponseWriter, r *http.Request)
	PutUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
}

type userController struct {
	db storage.DB
}

func NewUserController(db storage.DB) UserController {
	return &userController{db}
}

func (u *userController) PostUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("could not unmarshal PostUser request body\n%v", err)
		utils.SendError(w, "Could not parse body as JSON", http.StatusBadRequest)
		return
	}

	if user.Username == "" {
		utils.SendError(w, "Username required", http.StatusBadRequest)
		return
	}

	user.Password, err = hashAndSalt(user.Password)
	if err != nil {
		log.Printf("could not hash password\n%v", err)
		utils.SendError(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	err = u.db.CreateUser(&user)
	if err != nil {
		log.Printf("could not insert user %v into database\n%v", user, err)
		utils.SendError(w, "Error inserting user into database", http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, nil, http.StatusNoContent)
}

func (u *userController) GetUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		utils.SendError(w, "Username required", http.StatusBadRequest)
		return
	}

	user, err := u.db.GetUser(username)
	if storage.IsNotFound(err) {
		utils.SendError(w, fmt.Sprintf("User %v not found", username), http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("could not get user %v from the database\n%v", username, err)
		utils.SendError(w, "Error getting user from database", http.StatusInternalServerError)
		return
	}

	user.Password = ""

	utils.SendSuccess(w, user, http.StatusOK)
}

func (u *userController) GetUserExists(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		utils.SendError(w, "Username required", http.StatusBadRequest)
		return
	}

	_, err := u.db.GetUser(username)
	if storage.IsNotFound(err) {
		utils.SendSuccess(w, map[string]bool{"exists": false}, http.StatusOK)
		return
	} else if err != nil {
		log.Printf("could not get user %v from the database\n%v", username, err)
		utils.SendError(w, "Error looking for user", http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, map[string]bool{"exists": true}, http.StatusOK)
}

func (u *userController) PutUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		utils.SendError(w, "Username required", http.StatusBadRequest)
		return
	}

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("could not unmarshal PutUser request body\n%v", err)
		utils.SendError(w, "Could not parse body as JSON", http.StatusBadRequest)
		return
	}

	if len(user.Password) < passMinLength {
		utils.SendError(w, fmt.Sprintf("Password must be at least %v characters", passMinLength), http.StatusBadRequest)
		return
	}

	user.Password, err = hashAndSalt(user.Password)
	if err != nil {
		log.Printf("could not hash password\n%v", err)
		utils.SendError(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	err = u.db.UpdateUser(username, &user)
	if err != nil {
		log.Printf("could not update user %v\n%v", user, err)
		utils.SendError(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, nil, http.StatusNoContent)
}

func (u *userController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		utils.SendError(w, "Username required", http.StatusBadRequest)
		return
	}

	err := u.db.DeleteUser(username)
	if storage.IsNotFound(err) {
		utils.SendError(w, fmt.Sprintf("User %v not found", username), http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("could not delete user %v\n%v", username, err)
		utils.SendError(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, nil, http.StatusNoContent)
}

func hashAndSalt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
