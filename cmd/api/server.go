package main

import (
	"github.com/reqlane/github-releases-notifier/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
