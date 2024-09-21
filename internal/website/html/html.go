package html

import (
	"embed"
	"html/template"
)

//go:embed base/* driver/*
var fs embed.FS

var (
	Index  = template.Must(template.ParseFS(fs, "base/base.html", "driver/index.html"))
	SignUp = template.Must(template.ParseFS(fs, "base/base.html", "driver/sign-up.html"))
	SignIn = template.Must(template.ParseFS(fs, "base/base.html", "driver/sign-in.html"))
)
