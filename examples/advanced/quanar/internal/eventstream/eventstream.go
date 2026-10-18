// Package eventstream listens to a server-sent event stream: open, every named event, close.
package eventstream

import (
	"bufio"
	"context"
	"net/http"
	"strings"
)

// Follow returns when the stream ends or cannot be opened: the other side is gone.
func Follow(ctx context.Context, url string, onOpen func(), onEvent func(name string)) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "text/event-stream")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return
	}
	onOpen()
	name := "message"
	lines := bufio.NewScanner(res.Body)
	for lines.Scan() {
		line := lines.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			name = strings.TrimSpace(line[len("event:"):])
		case line == "":
			onEvent(name)
			name = "message"
		}
	}
}
