package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/rss-creator/controllers"
)

func Route(
	r *mux.Router,
	user controllers.UserController,
	auth controllers.AuthController) {

	r.HandleFunc("/health",
		GetHealth,
	).Methods(http.MethodGet)

	r.HandleFunc("/users",
		user.PostUser).Methods(http.MethodPost)
	r.HandleFunc("/users/exists",
		user.GetUserExists).Methods(http.MethodGet)
	r.HandleFunc("/users/{username}",
		auth.Wrapper(controllers.AccessTokenType, user.GetUser)).Methods(http.MethodGet)
	r.HandleFunc("/users/{username}",
		auth.Wrapper(controllers.AccessTokenType, user.PutUser)).Methods(http.MethodPut)
	r.HandleFunc("/users/{username}",
		auth.Wrapper(controllers.AccessTokenType, user.DeleteUser)).Methods(http.MethodDelete)

	r.HandleFunc("/users/{username}/authorize",
		auth.GetRefreshToken).Methods(http.MethodGet)
	r.HandleFunc("/users/{username}/token",
		auth.Wrapper(controllers.RefreshTokenType, auth.GetAuthToken)).Methods(http.MethodGet)
}

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
