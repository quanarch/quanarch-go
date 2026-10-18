// Command register seals every built quan: it hashes its program in the artifacts and writes the
// quanar's list. In a real system CI does this after uploading the program. With --open no hash
// is written, so a quan rebuilt by its watch is taken in as it is.
//
//	go run ./tools/cmd/register [--open]
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"shop-tools"
)

type source struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Hash   string   `json:"hash,omitempty"`
	Events string   `json:"events"`
	Quas   []string `json:"quas"`
}

func main() {
	log.SetFlags(0)
	open := flag.Bool("open", false, "write no hashes")
	flag.Parse()
	root := tools.Root()
	artifacts := tools.Env("ARTIFACTS_URL", "http://localhost:"+tools.Env("ARTIFACTS_PORT", "4100"))
	var list []source
	for _, name := range tools.Quans(root) {
		program, err := os.ReadFile(filepath.Join(root, "artifacts", "quans", name, name))
		if err != nil {
			log.Fatalf("%s is not built: %v", name, err)
		}
		s := source{Name: name, URL: fmt.Sprintf("%s/quans/%s/%s", artifacts, name, name), Events: artifacts + "/events/" + name, Quas: []string{"shop-qua"}}
		sum := sha256.Sum256(program)
		hash := hex.EncodeToString(sum[:])
		if !*open {
			s.Hash = hash
		}
		list = append(list, s)
		shown := "no hash"
		if s.Hash != "" {
			shown = hash[:12] + "…"
		}
		fmt.Printf("%-10s %s\n", name, shown)
	}
	b, _ := json.MarshalIndent(list, "", "  ")
	target := filepath.Join(root, "quanar", "data", "quans.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(target, append(b, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}
