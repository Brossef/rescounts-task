package main

import (
	"net/http"
)

func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user_id from context
		uidVal := r.Context().Value("user_id")
		if uidVal == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID, ok := uidVal.(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user_id exists in `admins` table
		var exists bool
		err := db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM admins WHERE user_id = $1);`,
			userID,
		).Scan(&exists)

		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Forbidden – admin only", http.StatusForbidden)
			return
		}

		// pass along to next handler
		next.ServeHTTP(w, r)
	})
}
