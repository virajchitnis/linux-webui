//go:build !production

package main

import "net/http"

// In dev mode the frontend is served by the Vite dev server.
func frontendHandler() http.Handler { return nil }
func hasFrontend() bool             { return false }
