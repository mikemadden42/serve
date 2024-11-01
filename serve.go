// http://stackoverflow.com/questions/11123865/golang-format-a-string-without-printing
// http://stackoverflow.com/questions/18537257/golang-how-to-get-the-directory-of-the-currently-running-file
// http://stackoverflow.com/questions/34017342/log-404-on-http-fileserver

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	portPtr := flag.Int("port", 8080, "http port")
	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	port := *portPtr
	addr := fmt.Sprintf(":%v", port)

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Serving %v on port %v\n", pwd, port)

	// Define the server with timeouts
	server := &http.Server{
		Addr:         addr,
		Handler:      wrapHandler(http.FileServer(http.Dir(pwd))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server with log.Fatal for error handling
	log.Fatal(server.ListenAndServe())
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
		srw := &statusRespWr{ResponseWriter: w}
		userAgent := r.Header.Get("User-Agent")
		h.ServeHTTP(srw, r)
		log.Printf("%s - [%s %s %s] %d [%s]", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, srw.status, userAgent)
	}
}
