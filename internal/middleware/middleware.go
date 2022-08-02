package middleware

import (
	"compress/gzip"
	"errors"
	"net/http"
	"strings"
)

type appHandler func(w http.ResponseWriter, r *http.Request) ([]byte, error)

func Middleware(h appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			defer gz.Close()
		}
		var appErr *AppError
		body, err := h(w, r)
		if err != nil {
			if errors.As(err, &appErr) {
				if errors.Is(err, ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					_, err := w.Write(ErrNotFound.Marshal())
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				}
				err = err.(*AppError)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(ErrNotFound.Marshal())
			}
			w.WriteHeader(http.StatusTeapot)
			w.Write(systemError(err).Marshal())
		}
		if body != nil {
			if strings.Contains(r.Header.Get(`Accept-Encoding`), `gzip`) {
				gz := gzip.NewWriter(w)
				defer gz.Close()
				w.Header().Set("Content-Encoding", "gzip")
				gz.Write(body)

			} else {
				w.Write(body)
			}
		}
	}
}
