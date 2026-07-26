package handler

import (
	"net/http"
	"strings"
	"net"
	"context"
	"github.com/google/uuid"
)

type RequestState struct {
	Err error
	UserID uuid.UUID
}

type contextKey string

const requestStateKey contextKey = "requestState"

type rwWrapper struct {
	rw http.ResponseWriter
	statusCode int
}

func NewRwWrapper(rw http.ResponseWriter) *rwWrapper {
	return &rwWrapper{
		rw: rw,
		statusCode: http.StatusOK,
	}
}

func (w *rwWrapper) Header() http.Header {
	return w.rw.Header()
}

func (w *rwWrapper) Write(b []byte) (int, error) {
	return w.rw.Write(b)
}

func (w *rwWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.rw.WriteHeader(statusCode)
}

const maxBytesSize = 1 << 20 // 1MB

func (h *Handler) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqURL := r.URL.String()

		hostIP := h.getClientIP(r)

		r.Body = http.MaxBytesReader(w, r.Body, maxBytesSize)


		rww := NewRwWrapper(w)
		holder := &RequestState{}
		ctx := context.WithValue(r.Context(), requestStateKey, holder)

		next.ServeHTTP(rww, r.WithContext(ctx))

		statusCode := rww.statusCode
		err := holder.Err
		userID := holder.UserID.String()

		if statusCode >= 400 && statusCode < 500 {
			h.warnLog(r, reqURL, statusCode, userID, hostIP, err)
		} else if statusCode >= 500 {
			h.errorLog(r, reqURL, statusCode, userID, hostIP, err)
		}else {
			h.accessLog(r, reqURL, statusCode, userID, hostIP)
		}
	})
}



func (h *Handler) getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		splitIps := strings.Split(xff, ",")
		if len(splitIps) > 0 {
			return strings.TrimSpace(splitIps[0])
		}
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	return ip
}

func (h *Handler) accessLog(r *http.Request, reqURL string, statusCode int, userID string, hostIP string) {
	accessLogger := h.logger.With("type", "access")
	accessLogger.Info("Access log",
		"method", r.Method,
		"url", reqURL,
		"status", statusCode,
		"userID", userID,
		"hostIP", hostIP,
	)
}

func (h *Handler) warnLog(r *http.Request, reqURL string, statusCode int, userID string, hostIP string, warn error) {
	warnLogger := h.logger.With("type", "warn")
	warnLogger.Warn("Warn log",
		"method", r.Method,
		"url", reqURL,
		"status", statusCode,
		"userID", userID,
		"hostIP", hostIP,
		"warn", warn,
	)
}
func (h *Handler) errorLog(r *http.Request, reqURL string, statusCode int, userID string, hostIP string, err error) {
	errorLogger := h.logger.With("type", "error")
	errorLogger.Error("Error log",
		"method", r.Method,
		"url", reqURL,
		"status", statusCode,
		"userID", userID,
		"hostIP", hostIP,
		"error", err,
	)
}
