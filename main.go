package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/rss-creator/controllers"
	"github.com/rss-creator/server"
	"github.com/rss-creator/storage"
)

func main() {
	getConfig()

	databaseType := viper.GetString("database.type")
	databasePath := viper.GetString("database.path")

	port := viper.GetString("server.port")
	cert := viper.GetString("server.cert")
	key := viper.GetString("server.key")
	jwtSecret := viper.GetString("server.jwtSecret")
	allowedOrigins := viper.GetStringSlice("server.allowedOrigins")

	db, err := storage.GetDB(databaseType, databasePath)
	if err != nil {
		log.Fatal("error connecting to database\n%v", err)
	}

	r := mux.NewRouter()
	uc := controllers.NewUserController(db)
	ac := controllers.NewAuthController(db, jwtSecret)
	server.Route(r.PathPrefix("/v1").Subrouter(), uc, ac)

	log.Printf("Listening on port %v", port)
	log.Fatal(http.ListenAndServeTLS(":"+port, cert, key, corsMiddleware(r, allowedOrigins)))
}

func getConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %v \n", err))
	}
}

func contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func corsMiddleware(h http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if !contains(allowedOrigins, origin) {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Password")
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}
