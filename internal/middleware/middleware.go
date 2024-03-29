package middleware

import (
	"compress/gzip"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type appHandler func(context *gin.Context) ([]byte, error)

func Middleware(h appHandler) gin.HandlerFunc {
	return func(context *gin.Context) {
		r := context.Request
		w := context.Writer

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
		body, err := h(context)
		if err != nil {
			if errors.As(err, &appErr) {
				if errors.Is(err, ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					_, err := w.Write(ErrNotFound.Marshal())
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				} else if errors.Is(err, UnknownMetricName) {
					w.WriteHeader(http.StatusNotImplemented)
					_, err := w.Write(UnknownMetricName.Marshal())
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				} else if errors.Is(err, DisconnectDB) {
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write(DisconnectDB.Marshal())
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				}

				err = err.(*AppError)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(ErrNotFound.Marshal())
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(systemError(err).Marshal())
		} else if body != nil {
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
