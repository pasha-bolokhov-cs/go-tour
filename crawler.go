package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}


var list = make(map[string]int)
// lock for 'list'
var listLock = make(chan int, 1)
// lock for printing
var printLock = make(chan int, 1)
// completion channel
var waitLock = make(chan int, 1)


// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, outChan chan<- int) {
	// check the depth
	if depth <= 0 {
		outChan <- -1
		return
	}

	// check the list first
	<- listLock
	if _, ok := list[url]; ok {
		listLock <- 1
		outChan <- 0
		return
	} else {
		list[url] = 1
	}
	listLock <- 1

	// now fetch
	body, urls, err := fetcher.Fetch(url)
	<- printLock
	if err != nil {
		fmt.Println(err)
		printLock <- 1
		outChan <- -2
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	printLock <- 1

	var num = len(urls)
	var currChan = make(chan int, num)
	for _, u := range urls {
		go Crawl(u, depth - 1, fetcher, currChan)
	}
	
	// wait for threads to complete
	for k := 0; k < num; k++ {
		<- currChan
	}

	outChan <- 1
	return
}

func main() {
	listLock <- 1  // unlock 'list'
	printLock <- 1 // unlock printing
	Crawl("http://golang.org/", 4, fetcher, waitLock)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
