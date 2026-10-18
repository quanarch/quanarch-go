// Command start builds and seals the quans, then starts the artifacts store, the quanar and the
// web client. Ctrl+C stops all three.
//
//	go run ./tools/cmd/start
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"shop-tools"
)

func relay(name string, r io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	lines := bufio.NewScanner(r)
	for lines.Scan() {
		fmt.Printf("%-12s%s\n", "["+name+"]", lines.Text())
	}
}

func main() {
	log.SetFlags(0)
	root := tools.Root()
	build := exec.Command("go", "run", "./tools/cmd/build")
	build.Dir, build.Stdout, build.Stderr = root, os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		log.Fatal(err)
	}
	// The quanar is built once, so stopping it stops the program itself, not go run.
	quanar := filepath.Join(root, "quanar", "bin", "quanar")
	compile := exec.Command("go", "build", "-o", quanar, ".")
	compile.Dir, compile.Stdout, compile.Stderr = filepath.Join(root, "quanar"), os.Stdout, os.Stderr
	if err := compile.Run(); err != nil {
		log.Fatal(err)
	}
	artifacts := filepath.Join(root, "tools", "bin", "artifacts")
	compile = exec.Command("go", "build", "-o", artifacts, "./cmd/artifacts")
	compile.Dir, compile.Stdout, compile.Stderr = filepath.Join(root, "tools"), os.Stdout, os.Stderr
	if err := compile.Run(); err != nil {
		log.Fatal(err)
	}

	parts := []struct {
		name string
		args []string
		dir  string
	}{
		{"artifacts", []string{artifacts}, root},
		{"quanar", []string{quanar}, filepath.Join(root, "quanar")},
		{"web", []string{"npm", "run", "dev"}, filepath.Join(root, "web")},
	}
	var wg sync.WaitGroup
	var children []*exec.Cmd
	for _, p := range parts {
		cmd := exec.Command(p.args[0], p.args[1:]...)
		cmd.Dir = p.dir
		out, _ := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		if err := cmd.Start(); err != nil {
			log.Fatalf("%s: %v", p.name, err)
		}
		wg.Add(1)
		go relay(p.name, out, &wg)
		children = append(children, cmd)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	// The terminal sent Ctrl+C to every part as well; make sure each one hears it.
	for _, c := range children {
		_ = c.Process.Signal(os.Interrupt)
	}
	for _, c := range children {
		_ = c.Wait()
	}
	wg.Wait()
}
