// Command build builds every quan into the artifacts with the quanar's quan-build, then seals them.
//
//	go run ./tools/cmd/build
package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"shop-tools"
)

func run(dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = dir, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("%s: %v", dir, err)
	}
}

func main() {
	log.SetFlags(0)
	root := tools.Root()
	for _, name := range tools.Quans(root) {
		out := filepath.Join(root, "artifacts", "quans", name, name)
		run(filepath.Join(root, "quans", name), "go", "tool", "quan-build", "--out", out)
	}
	run(root, "go", "run", "./tools/cmd/register")
}
