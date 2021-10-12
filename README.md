# Go cookies and JWT

Implementing a Go web app with [JWT](https://pkg.go.dev/github.com/golang-jwt/jwt) based authorisation via cookie storage for a private route.

This is not a test of secure authentication, encryption or data storage, so the page uses a very obvious password and no attempt to hash and store passwords.

## To run this project

1. Have Go installed on your system. This was made with v1.17. Grab [here](https://golang.org/) if needed.
2. Clone this repo to your system.
3. In the route folder run `go mod download` to get dependencies
4. run `go run main.go` to start the server.
5. Navigate to localhost:5000 in the browser to view the app.




