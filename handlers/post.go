package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"blog-apii/handlers/middleware"
	"blog-apii/models"
	"blog-apii/utils"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}
	if claims.Role != "user" {
		http.Error(w, "Forbidden: Invalid Role", http.StatusForbidden)
		return
	}

	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		fmt.Println("ERROR JSON ==", err) // Ubah dari "post" ke "err"
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	post.Status = "draft"
	post.PublishDate = time.Time{} // initialize PublishDate to zero value

	tags := make([]models.Tag, len(post.Tags))
	for i, val := range post.Tags {
		tags[i] = models.Tag{Label: val}
	}

	fmt.Println("TEXTTTT==", tags)

	db := utils.ConnectDB()
	defer db.Close()
	fmt.Println("tetete==")
	err = db.QueryRow("INSERT INTO posts (title, content, status) VALUES ($1, $2, $3) RETURNING id",
		post.Title, post.Content, post.Status).Scan(&post.ID)
	if err != nil {
		fmt.Println("Error", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	for _, tag := range post.Tags {
		_, err := db.Exec("INSERT INTO tags (label, posts_id) VALUES ($1, $2)", tag, post.ID)
		if err != nil {
			fmt.Println("Error inserting tag:", err)
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)

}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}
	if claims.Role != "user" && claims.Role != "admin" {
		http.Error(w, "Forbidden: Invalid Role", http.StatusForbidden)
		return
	}

	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		fmt.Println("ERROR JSON ==", err) // Ubah dari "post" ke "err"
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	post.Status = "draft"
	post.PublishDate = time.Time{} // initialize PublishDate to zero value

	tags := make([]models.Tag, len(post.Tags))
	for i, val := range post.Tags {
		tags[i] = models.Tag{Label: val}
	}

	fmt.Println("TEXTTTT==", tags)

	db := utils.ConnectDB()
	defer db.Close()
	fmt.Println("tetete==")
	err = db.QueryRow("UPDATE posts set title=$1, content=$2)",
		post.Title, post.Content, tags).Scan(&post.ID)
	if err != nil {
		fmt.Println("Error", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)

}

func PublishPost(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok || claims.Role != "admin" {
		http.Error(w, "Forbidden: Invalid Role", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	db := utils.ConnectDB()
	defer db.Close()

	_, err := db.Exec("UPDATE posts SET status='publish', publish_date=$1 WHERE id=$2", time.Now(), id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// w.WriteHeader(http.StatusOK)
	jsonResponse := map[string]string{"message": "Admin successfully to publish", "status": "success to publish"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonResponse)
}
