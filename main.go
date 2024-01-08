package main

import (
	"embed"

	"github.com/seanburman/kachekrow/pkg/store"
)

var (
	//go:embed client/web-build
	_ embed.FS
)

func main() {
	finish := make(chan bool)
	store.KacheKrow()
	<-finish
}
