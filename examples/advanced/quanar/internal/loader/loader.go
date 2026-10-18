// Package loader downloads a quan's file and checks it. Nothing runs here.
package loader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"shop-quanar/internal/config"
)

// ErrUnreachable means the quan's file could not be fetched.
var ErrUnreachable = errors.New("unreachable")

// ErrHashMismatch means the file is not the one the list expects.
var ErrHashMismatch = errors.New("hash-mismatch")

// Fetch downloads the quan's file.
func Fetch(ctx context.Context, source config.QuanSource) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cache-Control", "no-store")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s could not be reached", ErrUnreachable, source.URL)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s answered %d", ErrUnreachable, source.URL, res.StatusCode)
	}
	return io.ReadAll(res.Body)
}

// Verify returns the file's hash, or an error when the list expects another one.
func Verify(source config.QuanSource, code []byte) (string, error) {
	sum := sha256.Sum256(code)
	actual := hex.EncodeToString(sum[:])
	if source.Hash != "" && actual != source.Hash {
		return "", fmt.Errorf("%w: expected %s…, got %s…", ErrHashMismatch, source.Hash[:12], actual[:12])
	}
	return actual, nil
}
