package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var templates *template.Template

// The Claims struct is used as a template to store claim data both before signing and after parsing the JWT
type Claims struct {
	Username string
	jwt.StandardClaims
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.gohtml", nil)
}

func privateHandler(w http.ResponseWriter, r *http.Request) {
	// Receive the cookie containing the JWT
	c, err := r.Cookie("token")

	// If there's no cookie you can't be authorised
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Extract the JWT string from the cookie and initialise a new Claims struct to receive the data
	tokenString := c.Value
	claims := &Claims{}

	//  Parse the JWT and handle errors that could occur during validation
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWTKEY")), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Beyond here the user is authorised, so we could make database or api calls to pass into the template.
	// User data that was stored in the claims e.g. admin=true can be used to dynamically display pages.
	// If the token is valid, show the private route page and pass in the Username that was stored in the JWT.
	templates.ExecuteTemplate(w, "private.gohtml", claims.Username)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// If this handler was called with "POST"
	if r.Method == "POST" {

		// Check the password they gave against the super secret password
		if r.FormValue("password") == os.Getenv("PASSWORD") {
			fmt.Println("Logging in:", r.FormValue("username"))

			// Create the claims that will be given to the JWT
			tokenExpires := time.Now().Add(15 * time.Minute)
			claims := &Claims{
				Username: r.FormValue("username"),
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: tokenExpires.Unix(),
				},
			}

			// Create then sign the JWT with the claims and JWT key
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(os.Getenv("JWTKEY")))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Set the cookie containing the JWT to expire at the same time as the token
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    tokenString,
				Secure:   true,
				HttpOnly: true,
				Expires:  tokenExpires,
				Path:     "/",
			})
			// ...and redirect to the private route
			http.Redirect(w, r, "/private", http.StatusSeeOther)

			// If their password was wrong they're unauthorised
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}

		// If the login request was not called with "POST"...
	} else {
		// Check for a token cookie
		_, err := r.Cookie("token")
		if err != nil {
			// If there's no existing token to check we can show the login page
			templates.ExecuteTemplate(w, "login.gohtml", nil)
		} else {
			// If there is a token we can divert to the private route handler for validation
			http.Redirect(w, r, "/private", http.StatusSeeOther)
		}
	}
}

// Removes the cookie containing the JWT and redirect away from the private route
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Secure:   true,
		HttpOnly: true,
		MaxAge:   -1,
		Path:     "/",
	})
	fmt.Println("Logging out:", r.FormValue("username"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// In reality these would not be stored where they can be seen on the client
	os.Setenv("PASSWORD", "secret")
	os.Setenv("JWTKEY", "chillidogs")

	fs := http.FileServer(http.Dir("static"))
	templates = template.Must(template.ParseGlob("templates/*.gohtml"))

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/private", privateHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)

	fmt.Printf("Listening on %v\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
