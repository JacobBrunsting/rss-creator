package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/rss-creator/storage"
	"github.com/rss-creator/utils"
)

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"

	refreshExpiryTime = (time.Minute * 60 * 24) * 60
	accessExpiryTime  = time.Minute * 60 * 2
)

type handler func(w http.ResponseWriter, r *http.Request)

type AuthController interface {
	GetRefreshToken(w http.ResponseWriter, r *http.Request)
	GetAuthToken(w http.ResponseWriter, r *http.Request)
	Wrapper(tokenType string, h handler) handler
}

type authController struct {
	db        storage.DB
	jwtSecret string
}

type tokens struct {
	RefreshToken string `json:"refreshToken,omitempty"`
	AccessToken  string `json:"accessToken"`
}

type TokenClaims struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	jwt.StandardClaims
}

func NewAuthController(db storage.DB, jwtSecret string) AuthController {
	return &authController{db, jwtSecret}
}

func (a *authController) GetRefreshToken(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		utils.SendError(w, "Username required", http.StatusBadRequest)
		return
	}

	password := r.Header.Get("Password")
	if password == "" {
		utils.SendError(w, "Password header required", http.StatusBadRequest)
		return
	}

	user, err := a.db.GetUser(username)
	if storage.IsNotFound(err) {
		utils.SendError(w, fmt.Sprintf("User %v not found", username), http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("could not get user %v from the database\n%v", username, err)
		utils.SendError(w, "Error getting user from database", http.StatusInternalServerError)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		utils.SendError(w, "Password incorrect", http.StatusUnauthorized)
		return
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"exp":      time.Now().Add(refreshExpiryTime).Unix(),
		"type":     RefreshTokenType,
		"username": username,
	}).SignedString([]byte(a.jwtSecret))
	if err != nil {
		log.Printf("could not generate refresh token\n%v", err)
		utils.SendError(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	accessToken, err := a.getAccessToken(username)
	if err != nil {
		log.Printf("could not generate access token\n%v", err)
		utils.SendError(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	t := tokens{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}

	utils.SendSuccess(w, t, http.StatusOK)

	err = a.db.UpdateTokenValidity(username, true)
	if err != nil {
		log.Printf("could not set user token validity to valid\n%v", err)
	}
}

func (a *authController) GetAuthToken(w http.ResponseWriter, r *http.Request) {
	var bearerToken string
	bearerTokens, ok := r.Header["Authorization"]
	if ok && len(bearerTokens) >= 1 {
		bearerToken = strings.TrimPrefix(bearerTokens[0], "Bearer ")
	}

	if bearerToken == "" {
		utils.SendError(w, "Bearer token required", http.StatusUnauthorized)
		return
	}

	username := mux.Vars(r)["username"]
	if username == "" {
		utils.SendError(w, "Username required", http.StatusBadRequest)
		return
	}

	user, err := a.db.GetUser(username)
	if storage.IsNotFound(err) {
		utils.SendError(w, fmt.Sprintf("User %v not found", username), http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("could not get user %v from the database\n%v", username, err)
		utils.SendError(w, "Error getting user from database", http.StatusInternalServerError)
		return
	}

	if user.InvalidatedTokens {
		utils.SendError(w, "Refresh token has been invalidated", http.StatusUnauthorized)
		return
	}

	if err := a.validateAccessToken(bearerToken, user.Username, RefreshTokenType); err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, err := a.getAccessToken(username)
	if err != nil {
		log.Printf("could not generate access token\n%v", err)
		utils.SendError(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, tokens{AccessToken: accessToken}, http.StatusOK)
}

func (a *authController) Wrapper(tokenType string, h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		var bearerToken string
		bearerTokens, ok := r.Header["Authorization"]
		if ok && len(bearerTokens) >= 1 {
			bearerToken = strings.TrimPrefix(bearerTokens[0], "Bearer ")
		}

		if bearerToken == "" {
			utils.SendError(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		username := mux.Vars(r)["username"]
		if username == "" {
			utils.SendError(w, "Username required", http.StatusBadRequest)
			return
		}

		err := a.validateAccessToken(bearerToken, username, tokenType)
		if err != nil {
			utils.SendError(w, err.Error(), http.StatusBadRequest)
			return
		}

		h(w, r)
		return
	}
}

func (a *authController) validateAccessToken(token string, username string, tokenType string) error {
	refreshToken, err := jwt.ParseWithClaims(token, &TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(a.jwtSecret), nil
		},
	)

	if claims, ok := refreshToken.Claims.(*TokenClaims); ok && refreshToken.Valid && err == nil {
		if claims.Type != tokenType {
			return fmt.Errorf("Invalid token provided, '%v' token expected, got token with type '%v'", tokenType, claims.Type)
		}
	} else {
		return fmt.Errorf("Invalid token")
	}

	return nil
}

func (a *authController) getAccessToken(username string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"exp":      time.Now().Add(accessExpiryTime).Unix(),
		"type":     AccessTokenType,
		"username": username,
	}).SignedString([]byte(a.jwtSecret))
}
