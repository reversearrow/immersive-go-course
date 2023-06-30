package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

var htmlOut = `
<!DOCTYPE html>
<html>
<em>Hello, world</em>
<p>Query parameters:
<ul>
%s
</ul>
`

func main() {
	err := godotenv.Load("locals.env")
	if err != nil {
		fmt.Printf("error loading environment file: %v", err)
		os.Exit(1)
	}

	limiter := rate.NewLimiter(100, 30)

	http.HandleFunc("/200", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("200"))
	})

	http.HandleFunc("/500", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Internal Server Error"))
	})

	http.Handle("/404", http.NotFoundHandler())

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(htmlOut))
	})

	http.HandleFunc("/limited", func(w http.ResponseWriter, req *http.Request) {
		if limiter.Allow() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/authenticate", func(w http.ResponseWriter, req *http.Request) {
		localUser := os.Getenv("AUTH_USERNAME")
		localPassword := os.Getenv("AUTH_PASSWORD")
		username, password, ok := req.BasicAuth()
		if !ok {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		if username == localUser && password == localPassword {
			http.Error(w, "username or password invalid", http.StatusUnauthorized)
			return
		}

		output := `
		<!DOCTYPE html>
		<html>
		Hello %s
		</html>
		`

		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(output, username)))
	})

	http.ListenAndServe(":8080", nil)
}
