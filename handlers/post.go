package handlers

import (
	"database/sql"
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
		fmt.Println("ERROR JSON ==", err)
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	post.Status = "draft"
	post.PublishDate = time.Time{}

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
	if claims.Role != "user" {
		http.Error(w, "Forbidden: Invalid Role", http.StatusForbidden)
		return
	}

	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		fmt.Println("ERROR JSON ==", err)
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	if post.Status != "draft" {
		fmt.Println("ERROR JSON ==", err)
		http.Error(w, "Forbidden: Only posts with status 'draft' can be updated", http.StatusForbidden)
		return
	}

	db := utils.ConnectDB()
	defer db.Close()

	_, err = db.Exec("UPDATE posts SET title=$1, content=$2 WHERE id=$3",
		post.Title, post.Content, post.ID)
	if err != nil {
		fmt.Println("Error updating post:", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM tags WHERE posts_id=$1", post.ID)
	if err != nil {
		fmt.Println("Error deleting tags:", err)
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
	fmt.Fprintf(w, "Post successfully updated")
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

func DeletePost(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Bad Request: Missing post ID", http.StatusBadRequest)
		return
	}

	db := utils.ConnectDB()
	defer db.Close()

	var status string
	err := db.QueryRow("SELECT status FROM posts WHERE id = $1", id).Scan(&status)
	if err != nil {
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	if claims.Role != "admin" && status != "draft" {
		http.Error(w, "Forbidden: Only posts with status 'draft' can be deleted by users", http.StatusForbidden)
		return
	}

	_, err = db.Exec("DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Post successfully deleted")
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}

	db := utils.ConnectDB()
	defer db.Close()

	var rows *sql.Rows
	var err error

	if claims.Role == "admin" {
		rows, err = db.Query("SELECT id, title, content, status, publish_date FROM posts")
		if err != nil {
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}
	} else {
		rows, err = db.Query("SELECT id, title, content, status, publish_date FROM posts WHERE owner_id = $1", claims.Subject)
		if err != nil {
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}
	}

	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Status, &post.PublishDate)
		if err != nil {
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}
