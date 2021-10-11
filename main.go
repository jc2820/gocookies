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

type Claims struct {
	Username string
	jwt.StandardClaims
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.gohtml", nil)
}

func privateHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenString := c.Value
	claims := &Claims{}
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
	fmt.Println(claims.Username)
	templates.ExecuteTemplate(w, "private.gohtml", claims.Username)

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		if r.FormValue("password") == os.Getenv("PASSWORD") {
			fmt.Println("Logging in:", r.FormValue("username"))

			tokenExpires := time.Now().Add(15 * time.Minute)
			claims := &Claims{
				Username: r.FormValue("username"),
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: tokenExpires.Unix(),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(os.Getenv("JWTKEY")))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    tokenString,
				Secure:   true,
				HttpOnly: true,
				Expires:  tokenExpires,
				Path:     "/",
			})
			http.Redirect(w, r, "/private", http.StatusSeeOther)
		} else {
			templates.ExecuteTemplate(w, "401.gohtml", nil)
		}
	} else {
		_, err := r.Cookie("token")
		if err != nil {
			templates.ExecuteTemplate(w, "login.gohtml", nil)
		} else {
			http.Redirect(w, r, "/private", http.StatusSeeOther)
		}
	}
}

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
