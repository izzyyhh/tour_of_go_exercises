package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type CrawlerMap struct {
	crawled map[string]string
	mu      sync.Mutex
}

func (m *CrawlerMap) Set(key string, value string) {
	m.mu.Lock()
	m.crawled[key] = value
	m.mu.Unlock()
}

func (m *CrawlerMap) Get(key string) (string, bool) {
	m.mu.Lock()
	val, ok := m.crawled[key]
	m.mu.Unlock()

	return val, ok
}

func Crawl(url string, depth int, fetcher Fetcher, cm *CrawlerMap, wg *sync.WaitGroup) {
	defer wg.Done()

	if depth <= 0 {
		return
	}

	_, alreadyCrawled := cm.Get(url)

	if alreadyCrawled {
		return
	}

	body, urls, err := fetcher.Fetch(url)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("found: %s %q\n", url, body)

	for _, u := range urls {
		wg.Add(1)
		go Crawl(u, depth-1, fetcher, cm, wg)
	}
}

func main() {
	wg := &sync.WaitGroup{}

	cm := &CrawlerMap{
		crawled: make(map[string]string),
	}

	Crawl("https://golang.org/", 4, fetcher, cm, wg)
	wg.Wait()
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
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
