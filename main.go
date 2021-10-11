package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var templates *template.Template

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.gohtml", nil)
}

func privateHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("user")
	if err != nil {
		templates.ExecuteTemplate(w, "401.gohtml", nil)
	} else {
		templates.ExecuteTemplate(w, "private.gohtml", nil)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.FormValue("password") == os.Getenv("PASSWORD") {
			fmt.Println("Logging in:", r.FormValue("username"))
			http.SetCookie(w, &http.Cookie{
				Name:     "user",
				Value:    r.FormValue("username"),
				Secure:   true,
				HttpOnly: true,
				MaxAge:   100,
				Path:     "/",
			})
			http.Redirect(w, r, "/private", http.StatusSeeOther)
		} else {
			templates.ExecuteTemplate(w, "401.gohtml", nil)
		}
	} else {
		_, err := r.Cookie("user")
		if err != nil {
			templates.ExecuteTemplate(w, "login.gohtml", nil)
		} else {
			http.Redirect(w, r, "/private", http.StatusSeeOther)
		}
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Secure:   true,
		HttpOnly: true,
		MaxAge:   -1,
		Path:     "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	os.Setenv("PASSWORD", "secret")

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
