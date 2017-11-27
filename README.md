# Robots Parser [![Build Status](https://travis-ci.org/samclarke/robotstxt.svg?branch=master)](https://travis-ci.org/samclarke/robotstxt) [![Coverage Status](https://coveralls.io/repos/github/samclarke/robotstxt/badge.svg?branch=master)](https://coveralls.io/github/samclarke/robotstxt?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/samclarke/robotstxt)](https://goreportcard.com/report/github.com/samclarke/robotstxt) [![GoDoc](https://godoc.org/github.com/samclarke/robotstxt?status.svg)](http://godoc.org/github.com/samclarke/robotstxt)

A robots.txt parser written in Go, based on the Node.js [robots-parser package](https://github.com/samclarke/robots-parser).

It currently supports:

  * User-agent:
  * Allow:
  * Disallow:
  * Sitemap:
  * Crawl-delay:
  * Host:
  * URL encoded & UTF-8 paths
  * Paths with wildcards (*) and EOL matching ($)

## Installation

Go get:

    go get github.com/samclarke/robotstxt

## Usage

```go
import (
    "log"
    "github.com/samclarke/robotstxt"
)

func main() {
    url := "http://www.example.com/robots.txt"
    contents := `
        User-agent: *
        Disallow: /dir/
        Disallow: /test.html
        Allow: /dir/test.html
        Allow: /test.html
        Crawl-delay: 1
        Sitemap: http://example.com/sitemap.xml
        Host: example.com
    `

    robots, err := Parse(contents, url)
    if err != nil {
        log.Fatalln(err.Error())
    }

    allowed, _ := robots.IsAllowed("Sams-Bot/1.0", "http://www.example.com/test.html")
    if !allowed {
        println("Not allowed to crawl: /test.html")
    }

    allowed, _ := robots.IsAllowed("Sams-Bot/1.0", "http://www.example.com/dir/test.html")
    if allowed {
        println("Allowed to crawl: /dir/test.html")
    }

    // 1
    println("Crawl delay: " + robots.CrawlDelay("Sams-Bot/1.0"))

    // [http://example.com/sitemap.xml]
    println("Sitemaps: " + strings.Join(robots.Sitemaps(), ","))

    // example.com
    println("Preferred host: " + robots.Host())
}
```

# License

	The MIT License (MIT)

	Copyright (c) 2017 Sam Clarke

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.
