package main

import (
	"fmt"
	"log"
	"mime"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mimes := map[string]string{
		".test":   "text/plain",
		".svelte": "text/html",
	}

	for ext, mimeType := range mimes {
		mime.AddExtensionType(ext, mimeType)
	}

	fmt.Println(mime.TypeByExtension(".svelte"))
	return nil
}
