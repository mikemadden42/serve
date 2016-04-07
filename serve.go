// http://golang.org/pkg/net/http/#example_FileServer
// http://stackoverflow.com/questions/11123865/golang-format-a-string-without-printing
// http://stackoverflow.com/questions/18537257/golang-how-to-get-the-directory-of-the-currently-running-file

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

	fmt.Printf("Serving %v on port %v", pwd, port)

	// Simple static webserver:
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir(pwd))))
}
