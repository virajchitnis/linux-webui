package middleware

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// No nonce: the embedded index.html is a static file, so we cannot inject
		// a per-request nonce into the script tag. Use 'self' to allow same-origin
		// scripts (the Vite bundle) and 'unsafe-inline' for Tailwind's style tags.
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' wss:; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
