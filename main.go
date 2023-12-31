package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	handlers "anidex_api/api/handlers"
	"anidex_api/api/helpers"
	middleware "anidex_api/api/middleware"
	DB "anidex_api/db"
	"anidex_api/http/responses"
)

func getStatus(w http.ResponseWriter, r *http.Request) {
	// Check if it's an OPTIONS request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Access-Control-Allow-Headers", "Authorization")
	// verify token
	authHeader := r.Header.Get("Authorization")

	// Check if the "Authorization" header is set
	if authHeader == "" {
		// Handle the case where the header is not provided
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	_, e := helpers.VerifyToken(authHeader)
	if e != nil {
		res, err := responses.CustomResponse(w, nil, e.Error(), http.StatusUnauthorized, e.Error())
		if err != nil {
			http.Error(w, e.Error(), http.StatusUnauthorized)
			return
		}
		w.Write(res)
		return
	}
	fmt.Println("OK")
}

func getFiles(w http.ResponseWriter, r *http.Request) {
	// Check if it's an OPTIONS request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")

	// verify token
	authHeader := r.Header.Get("Authorization")

	// Check if the "Authorization" header is set
	if authHeader == "" {
		// Handle the case where the header is not provided
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	_, e := helpers.VerifyToken(authHeader)
	if e != nil {
		res, err := responses.CustomResponse(w, nil, e.Error(), http.StatusUnauthorized, e.Error())
		if err != nil {
			http.Error(w, e.Error(), http.StatusUnauthorized)
			return
		}
		w.Write(res)
		return
	}

	entries, err := os.ReadDir(os.Getenv("RAILWAY_VOLUME_MOUNT_PATH") + "/uploaded_images")
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		fmt.Println(e.Name())
	}
}

func removeFiles(w http.ResponseWriter, r *http.Request) {
	// Check if it's an OPTIONS request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")

	// verify token
	authHeader := r.Header.Get("Authorization")

	// Check if the "Authorization" header is set
	if authHeader == "" {
		// Handle the case where the header is not provided
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	_, e := helpers.VerifyToken(authHeader)
	if e != nil {
		res, err := responses.CustomResponse(w, nil, e.Error(), http.StatusUnauthorized, e.Error())
		if err != nil {
			http.Error(w, e.Error(), http.StatusUnauthorized)
			return
		}
		w.Write(res)
		return
	}

	entries, err := os.ReadDir(os.Getenv("RAILWAY_VOLUME_MOUNT_PATH") + "/uploaded_images")
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		err := os.Remove(os.Getenv("RAILWAY_VOLUME_MOUNT_PATH") + "/uploaded_images/" + e.Name())
		if err != nil {
			fmt.Println(os.Getenv("RAILWAY_VOLUME_MOUNT_PATH")+"/uploaded_images/"+e.Name()+"could not be deleted", err.Error())
		}
	}
}

func setupRoutes(port string, db *sql.DB) {
	if port == "" {
		port = "3000"
	}

	if db == nil { //test without DB

		//Animal
		http.HandleFunc("/animal", handlers.CreateAnimal)
		http.HandleFunc("/animals/category/{category}/page/{page}", handlers.GetAnimals)

		//Login
		http.HandleFunc("/login", handlers.Login)

	} else {
		//Animal
		http.HandleFunc("/animal", middleware.WithDatabase(db, handlers.CUDAnimal))
		http.HandleFunc("/animals", middleware.WithDatabase(db, handlers.GetAnimals))

		//Images
		http.HandleFunc("/images", handlers.GetImageByPath)

		//Category
		http.HandleFunc("/categories", middleware.WithDatabase(db, handlers.GetCategories))

		//Stats
		http.HandleFunc("/stats", middleware.WithDatabase(db, handlers.GetStats))

		//Login
		http.HandleFunc("/login", middleware.WithDatabase(db, handlers.Login))
	}

	http.HandleFunc("/", getStatus)
	http.HandleFunc("/getFiles", getFiles)
	http.HandleFunc("/remove", removeFiles)
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}

func main() {

	port := os.Getenv("PORT")

	// Initialize DB
	if port != "" {
		db, err := DB.InitializeDB()
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("Connection with DB established!")
		}
		defer db.Close()
		fmt.Println("Server online reachable at port", port)
		setupRoutes(port, db)
	} else {
		fmt.Println("Server online reachable at port 3000")
		setupRoutes(port, nil)
	}

}
