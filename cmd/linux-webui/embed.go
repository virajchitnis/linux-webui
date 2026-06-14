//go:build production

package main

import (
	"net/http"

	"github.com/virajchitnis/linux-webui/ui"
)

func frontendHandler() http.Handler { return ui.Handler() }
