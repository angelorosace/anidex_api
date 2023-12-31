package handlers

import (
	responses "anidex_api/http/responses"
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID       int    `json:"id"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	//get password from url
	//get username from url
	var credentials Credentials
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	if username == "" || password == "" {
		resp, err := responses.MissingURLParametersResponse(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(resp)
		return
	}

	credentials.UserName = username
	credentials.Password = password

	//get user with same username from DB
	db := r.Context().Value("db").(*sql.DB)

	// Query the database
	query := "SELECT * FROM users WHERE username = ?"
	rows, err := db.Query(query, credentials.UserName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var user User
	for rows.Next() {
		err := rows.Scan(
			&user.ID,
			&user.UserName,
			&user.Password,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if user.UserName == "" {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	//get salt from server
	salt := os.Getenv("SALT")

	//add salt to password
	passwordBytes := []byte(credentials.Password)
	passwordBytes = append(passwordBytes, []byte(salt)...)

	// Compare the stored hashed password, with the hashed version of the password that was received
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), passwordBytes); err != nil {
		// If the two passwords don't match, return a 401 status
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	//create JWT token
	expiresAt := time.Now().Add(time.Hour * 24) // Token expiration time (adjust as needed)

	claims := jwt.MapClaims{}
	claims["username"] = credentials.UserName
	claims["exp"] = expiresAt.Unix()

	//send back JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(salt))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, e := responses.CustomResponse(w, tokenString, "User Authenticated", http.StatusOK, "")
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
