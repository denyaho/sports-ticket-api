package handler

import (
	"net/http"
	"io"
	"bytes"
	"strings"
	"net"
	"log/slog"
)
type contextKey string

const UserIDKey contextKey = "userID"

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
	return
}

type LoggingBundle struct {
	logger *slog.Logger
}


func (h *Handler) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqURL := r.URL.String()
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.errorLog(r, "",http.StatusInternalServerError, "", "", nil, err)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		hostIP := h.getClientIP(r)
		rww := NewRwWrapper(w)
		next.ServeHTTP(rww, r)
		statusCode := rww.statusCode
		userID, _ := r.Context().Value(UserIDKey).(string)

		h.accessLog(r, reqURL, statusCode, userID, hostIP, reqBody)
		if statusCode >= 500 {
			h.errorLog(r, reqURL, statusCode, userID, hostIP, reqBody, err)
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

func (h *Handler) accessLog(r *http.Request, reqURL string, statusCode int, userID string, hostIP string, reqBody []byte) {
	accessLogger := h.logger.With("type", "access")
	accessLogger.Info("Access log",
		"method", r.Method,
		"url", reqURL,
		"status", statusCode,
		"userID", userID,
		"hostIP", hostIP,
		"requestBody", string(reqBody),
	)
}

func (h *Handler) errorLog(r *http.Request, reqURL string, statusCode int, userID string, hostIP string, reqBody []byte, err error) {
	errorLogger := h.logger.With("type", "error")
	errorLogger.Error("Error log",
		"url", reqURL,
		"status", statusCode,
		"userID", userID,
		"requestBody", string(reqBody),
		"error", err.Error(),
	)
}
