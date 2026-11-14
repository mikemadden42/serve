// Command serve is a simple HTTP file server for the current directory.
//
// It listens on 127.0.0.1:8080 by default, hides dotfiles, and logs each
// request. Use -bind 0.0.0.0 to serve on all interfaces.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	portPtr := flag.Int("port", 8080, "http port")
	bindPtr := flag.String("bind", "127.0.0.1", "address to listen on (use 0.0.0.0 for all interfaces)")
	flag.Parse()

	port := *portPtr
	addr := net.JoinHostPort(*bindPtr, strconv.Itoa(port))

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Serving %v on http://%v\n", pwd, addr)

	// Define the server with timeouts
	server := &http.Server{
		Addr:        addr,
		Handler:     wrapHandler(http.FileServer(dotFileHidingFS{http.Dir(pwd)})),
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	// Start the server with log.Fatal for error handling
	log.Fatal(server.ListenAndServe())
}

// containsDotFile reports whether any element of name begins with a period.
func containsDotFile(name string) bool {
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

// dotFileHidingFS is an http.FileSystem that hides dotfiles and dot directories.
type dotFileHidingFS struct {
	http.FileSystem
}

// Open returns fs.ErrNotExist for any path containing a dot element, so the
// file server responds with 404.
func (fsys dotFileHidingFS) Open(name string) (http.File, error) {
	if containsDotFile(name) {
		return nil, fs.ErrNotExist
	}
	f, err := fsys.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return dotFileHidingFile{f}, nil
}

// dotFileHidingFile is an http.File that omits dotfiles from directory listings.
type dotFileHidingFile struct {
	http.File
}

// Readdir filters dotfiles out of the underlying directory listing.
func (f dotFileHidingFile) Readdir(n int) ([]fs.FileInfo, error) {
	files, err := f.File.Readdir(n)
	visible := files[:0]
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			visible = append(visible, file)
		}
	}
	return visible, err
}

type statusRespWr struct {
	http.ResponseWriter
	status int
}

func (w *statusRespWr) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func wrapHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		srw := &statusRespWr{ResponseWriter: w, status: http.StatusOK}
		userAgent := r.Header.Get("User-Agent")
		h.ServeHTTP(srw, r)
		//nolint:gosec // G706: path and user agent are %q-quoted; method and proto are validated by net/http
		log.Printf("%s - [%s %q %s] %d [%q]", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, srw.status, userAgent)
	}
}
