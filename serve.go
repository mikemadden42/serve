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

	// https://golang.org/pkg/net/http/#example_FileServer
	//log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir(pwd))))
	log.Fatal(http.ListenAndServe(addr, wrapHandler(http.FileServer(http.Dir(pwd)))))

	// https://golang.org/pkg/net/http/#example_FileServer_stripPrefix
	//http.HandleFunc("/o/", wrapHandler(http.StripPrefix("/o", http.FileServer(http.Dir(pwd)))))
	//panic(http.ListenAndServe(addr, nil))
}

type StatusRespWr struct {
	http.ResponseWriter
	status int
}

func (w *StatusRespWr) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func wrapHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		srw := &StatusRespWr{ResponseWriter: w}
		h.ServeHTTP(srw, r)
		log.Printf("-> status code: %d, path: %s, host: %s", srw.status, r.RequestURI, r.RemoteAddr)
	}
}
