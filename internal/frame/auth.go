// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frame

import (
	"net/http"
	"strings"
)

// authMiddleware enforces the shared-secret bearer token, mirroring
// ImmichFrame's ImmichFrameAuthenticationHandler. When secret is empty the API
// is open.
func authMiddleware(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeText(w, http.StatusUnauthorized, "Missing Authorization Header")
			return
		}
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			writeText(w, http.StatusUnauthorized, "Invalid Authorization Header")
			return
		}
		token := strings.TrimSpace(authHeader[len("Bearer "):])
		if token != secret {
			writeText(w, http.StatusUnauthorized, "The AuthenticationSecret was not correct!")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}
