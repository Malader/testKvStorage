package middleware

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(next http.Handler) http.Handler {
	return middleware.RequestLogger(
		&middleware.DefaultLogFormatter{
			Logger:  log.New(os.Stdout, "", log.LstdFlags),
			NoColor: false,
		},
	)(next)
}

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return middleware.Timeout(duration)
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}
