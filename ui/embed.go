//go:build production

package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var embeddedFS embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(embeddedFS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if file exists; if not, serve index.html for SPA routing.
		f, err := sub.Open(r.URL.Path)
		if err != nil {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}
