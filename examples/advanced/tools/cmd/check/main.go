// Command check vets every Go module of the shop and type-checks the web client.
//
//	go run ./tools/cmd/check
package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"shop-tools"
)

func main() {
	log.SetFlags(0)
	root := tools.Root()
	dirs := []string{"packages/shop-qua", "quanar", "tools"}
	for _, quan := range tools.Quans(root) {
		dirs = append(dirs, "quans/"+quan)
	}
	var failed []string
	for _, dir := range dirs {
		cmd := exec.Command("go", "vet", "./...")
		cmd.Dir, cmd.Stdout, cmd.Stderr = filepath.Join(root, dir), os.Stdout, os.Stderr
		if cmd.Run() != nil {
			failed = append(failed, dir)
		}
	}
	web := exec.Command("npm", "run", "typecheck")
	web.Dir, web.Stdout, web.Stderr = filepath.Join(root, "web"), os.Stdout, os.Stderr
	if web.Run() != nil {
		failed = append(failed, "web")
	}
	if len(failed) > 0 {
		log.Fatalf("failed: %v", failed)
	}
	log.Print("every part is clean")
}
