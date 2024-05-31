package main

import (
	"log"
	"net/http"

	"blog-apii/handlers"
	"blog-apii/handlers/middleware"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/register", handlers.Register).Methods("POST")

	postRouter := r.PathPrefix("/posts").Subrouter()
	postRouter.Use(middleware.AuthMiddleware)
	postRouter.HandleFunc("", handlers.CreatePost).Methods("POST")
	postRouter.HandleFunc("", handlers.UpdatePost).Methods("PUT")
	postRouter.HandleFunc("/publish/:id", handlers.PublishPost).Methods("PUT")

	log.Fatal(http.ListenAndServe(":9090", r))
}
