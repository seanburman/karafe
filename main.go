package main

import (
	"github.com/seanburman/kaw/pkg/store"
)

func main() {
	finish := make(chan bool)
	store.Kaw()

	<-finish
}
