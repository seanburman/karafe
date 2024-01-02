package main

import (
	"embed"

	"github.com/seanburman/kaw/pkg/store"
)

var (
	//go:embed dist
	_ embed.FS
)

func main() {
	finish := make(chan bool)
	store.Kaw()
	<-finish
}
