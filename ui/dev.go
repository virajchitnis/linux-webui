//go:build !production

package ui

import "net/http"

func Handler() http.Handler { return nil }
