// Command artifacts is a stand-in for S3 or a CDN: it serves the built quans, and announces a
// new build of a quan on /events/<quan> so a quanar following it picks the new program up at once.
//
//	go run ./tools/cmd/artifacts    port 4100, or ARTIFACTS_PORT
package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"shop-tools"
)

type streams struct {
	mu   sync.Mutex
	open map[string]map[chan struct{}]bool // quan → streams
}

func (s *streams) announce(quan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.open[quan] {
		select {
		case c <- struct{}{}:
		default:
		}
	}
}

func (s *streams) serve(w http.ResponseWriter, r *http.Request) {
	quan := r.PathValue("quan")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, _ := w.(http.Flusher)
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()
	c := make(chan struct{}, 1)
	s.mu.Lock()
	if s.open[quan] == nil {
		s.open[quan] = map[chan struct{}]bool{}
	}
	s.open[quan][c] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.open[quan], c)
		s.mu.Unlock()
	}()
	for {
		select {
		case <-c:
			fmt.Fprint(w, "event: change\ndata: {}\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func main() {
	log.SetFlags(0)
	root := filepath.Join(tools.Root(), "artifacts", "quans")
	if err := os.MkdirAll(root, 0o755); err != nil {
		log.Fatal(err)
	}
	s := &streams{open: map[string]map[chan struct{}]bool{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /events/{quan}", s.serve)
	mux.Handle("GET /quans/", http.StripPrefix("/quans/", http.FileServer(http.Dir(root))))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	watched := map[string]bool{}
	addAll := func() {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.IsDir() && !watched[p] {
				watched[p] = true
				_ = watcher.Add(p)
			}
			return nil
		})
	}
	addAll()
	go func() {
		// A build writes in bursts; announce once it settles.
		pending := map[string]bool{}
		var settle <-chan time.Time
		for {
			select {
			case e := <-watcher.Events:
				addAll()
				if rel, err := filepath.Rel(root, e.Name); err == nil && !strings.HasPrefix(rel, "..") && !strings.HasSuffix(e.Name, ".tmp") {
					pending[strings.Split(filepath.ToSlash(rel), "/")[0]] = true
					settle = time.After(150 * time.Millisecond)
				}
			case <-settle:
				for quan := range pending {
					s.announce(quan)
				}
				pending = map[string]bool{}
			case err := <-watcher.Errors:
				log.Print(err)
			}
		}
	}()
	port := tools.Env("ARTIFACTS_PORT", "4100")
	log.Printf("artifacts on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
