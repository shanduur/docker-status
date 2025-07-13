package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var static embed.FS

var UiHandler http.Handler = func() http.Handler {
	index, _ := fs.Sub(static, "static")
	return http.FileServer(http.FS(index))
}()
