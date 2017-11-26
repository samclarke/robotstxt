package robotstxt

import (
	"reflect"
	"testing"
	"time"
)

func testRobots(t *testing.T, contents string, url string, allowed []string, disallowed []string) {
	robots, _ := Parse(contents, url)

	for _, path := range allowed {
		allowed, err := robots.IsAllowed("*", path)
		if err != nil {
			t.Errorf(err.Error())
		} else if !allowed {
			t.Errorf("The path " + path + " should be allowed")
		}
	}

	for _, path := range disallowed {
		allowed, err := robots.IsAllowed("*", path)
		if err != nil {
			t.Errorf(err.Error())
		} else if allowed {
			t.Errorf("The path " + path + " should be disallowed")
		}
	}
}

func TestRobotsTxt_prseTheDisallowDirective(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent: *
		Disallow: /fish/
		Disallow: /test.html
	`

	allowed := []string{
		"http://www.example.com/fish",
		"http://www.example.com/Test.html",
	}

	disallowed := []string{
		"http://www.example.com/fish/index.php",
		"http://www.example.com/fish/",
		"http://www.example.com/test.html",
	}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_parsePatterns(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent: *
		Disallow: /fish*.php
		Disallow: /*.dext$
	`

	allowed := []string{
		"http://www.example.com/Fish.PHP",
		"http://www.example.com/Fish.dext1",
	}

	disallowed := []string{
		"http://www.example.com/fish.php",
		"http://www.example.com/fishheads/catfish.php?parameters",
		"http://www.example.com/AnYthInG.dext",
		"http://www.example.com/Fish.dext.dext",
	}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_correctOrderPresidenceForAllowAndDisallow(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent: *
		Disallow: /fish*.php
		Allow: /fish/index.php
		Disallow: /test
		Allow: /test/
	`

	allowed := []string{
		"http://www.example.com/test/index.html",
		"http://www.example.com/test/",
	}

	disallowed := []string{
		"http://www.example.com/fish.php",
		"http://www.example.com/fishheads/catfish.php?parameters",
		"http://www.example.com/fish/index.php",
		"http://www.example.com/test",
	}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_ignoreRulesThatAreNotInAGroup(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		Disallow: /secret.html
		Disallow: /test
	`

	allowed := []string{
		"http://www.example.com/secret.html",
		"http://www.example.com/test/index.html",
		"http://www.example.com/test/",
	}

	disallowed := []string{}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_ignoreComments(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		#
		# This is a comment
		#
		User-agent: *
		# This is a comment
		Disallow: /fish/
		# Disallow: fish
		Disallow: /test.html
	`

	allowed := []string{
		"http://www.example.com/fish",
		"http://www.example.com/Test.html",
	}

	disallowed := []string{
		"http://www.example.com/fish/index.php",
		"http://www.example.com/fish/",
		"http://www.example.com/test.html",
	}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_ignoreInvalidLines(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		invalid line
		User-agent: *
		Disallow: /fish/
		:::::another invalid line:::::
		Disallow: /test.html
		Unknown: tule
	`

	allowed := []string{
		"http://www.example.com/fish",
		"http://www.example.com/Test.html",
	}

	disallowed := []string{
		"http://www.example.com/fish/index.php",
		"http://www.example.com/fish/",
		"http://www.example.com/test.html",
	}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_ignoreeEmptyUserAgentLines(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent:
		Disallow: /fish/
		Disallow: /test.html
	`

	allowed := []string{
		"http://www.example.com/fish",
		"http://www.example.com/Test.html",
		"http://www.example.com/fish/index.php",
		"http://www.example.com/fish/",
		"http://www.example.com/test.html",
	}

	disallowed := []string{}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_supportGroupsWithMultipleUserAgents(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent: agEnTa
		User-agent: agentb
		Disallow: /fish
	`

	robots, _ := Parse(contents, url)

	allowed, _ := robots.IsAllowed("agenta", "http://www.example.com/test.html")
	if !allowed {
		t.Errorf("The path /test.html should be allowed")
	}

	allowed, _ = robots.IsAllowed("agentb", "http://www.example.com/test.html")
	if !allowed {
		t.Errorf("The path /test.html should be allowed")
	}

	allowed, _ = robots.IsAllowed("agenta", "http://www.example.com/fish/test.html")
	if allowed {
		t.Errorf("The path /fish/test.html should be disallowed")
	}

	allowed, _ = robots.IsAllowed("agentb", "http://www.example.com/fish/test.html")
	if allowed {
		t.Errorf("The path /fish/test.html should be disallowed")
	}
}

func TestRobotsTxt_returnErrorForInvalidUrls(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent: *
		Disallow: /secret.html
		Disallow: /test
	`

	invalidUrls := []string{
		"http://example.com/secret.html",
		"http://www.example.net/test/index.html",
		"http://www.examsple.com/test/",
		"h:||@@##'#']s;a[//test/",
	}

	robots, _ := Parse(contents, url)

	for _, u := range invalidUrls {
		_, err := robots.IsAllowed("*", u)
		_, ok := err.(*InvalidHostError)
		if !ok {
			t.Errorf("The URL " + u + " should cause an error")
		}
	}
}

func TestRobotsTxt_handleUrlsWithPunycode(t *testing.T) {
	url := "http://www.münich.com/robots.txt"
	contents := `
		User-agent: *
		Disallow: /secret.html
		Disallow: /test
	`

	allowed := []string{
		"http://www.münich.com/index.html",
		"http://www.xn--mnich-kva.com/index.html",
	}

	disallowed := []string{
		"http://www.münich.com/secret.html",
		"http://www.xn--mnich-kva.com/secret.html",
	}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_allowAllIfEmptyRobotsTxt(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := ``

	allowed := []string{
		"http://www.example.com/secret.html",
		"http://www.example.com/test/index.html",
		"http://www.example.com/test/",
	}

	disallowed := []string{}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_parseTheCrawlDelayDirective(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: a
		crawl-delay: 1

		user-agent: b
		disallow: /d

		user-agent: c
		user-agent: d
		crawl-delay: 10
	`

	robots, _ := Parse(contents, url)

	if robots.CrawlDelay("a") != time.Second {
		t.Errorf("Expected crawl delay for a to be 1")
	}

	if robots.CrawlDelay("b") != 0 {
		t.Errorf("Expected crawl delay for b to be 0")
	}

	if robots.CrawlDelay("c") != 10*time.Second {
		t.Errorf("Expected crawl delay for c to be 10")
	}

	if robots.CrawlDelay("d") != 10*time.Second {
		t.Errorf("Expected crawl delay for d to be 10")
	}
}

func TestRobotsTxt_returnZeroIfEmpty(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := ``

	robots, _ := Parse(contents, url)

	if robots.CrawlDelay("a") != 0 {
		t.Errorf("Expected crawl delay for a to be 0")
	}
}

func TestRobotsTxt_ignoreInvalidCrawlDelayDirectives(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: a
		crawl-delay: 1.2.1

		user-agent: b
		crawl-delay: 1.a0

		user-agent: c
		user-agent: d
		crawl-delay: 10a
	`

	robots, _ := Parse(contents, url)

	if robots.CrawlDelay("a") != 0 {
		t.Errorf("Expected crawl delay for a to be 0")
	}

	if robots.CrawlDelay("b") != 0 {
		t.Errorf("Expected crawl delay for b to be 0")
	}

	if robots.CrawlDelay("c") != 0 {
		t.Errorf("Expected crawl delay for c to be 0")
	}

	if robots.CrawlDelay("d") != 0 {
		t.Errorf("Expected crawl delay for d to be 0")
	}
}

func TestRobotsTxt_parseTheSitemapDirective(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: a
		crawl-delay: 1
		sitemap: http://example.com/test.xml

		user-agent: b
		disallow: /d

		sitemap: /sitemap.xml
		sitemap:     http://example.com/test/sitemap.xml   
	`

	expected := []string{
		"http://example.com/test.xml",
		"/sitemap.xml",
		"http://example.com/test/sitemap.xml",
	}

	robots, _ := Parse(contents, url)

	if !reflect.DeepEqual(robots.Sitemaps(), expected) {
		t.Errorf("Expected sitemaps to match")
	}
}

func TestRobotsTxt_parseTheHostDirective(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: a
		crawl-delay: 1
		host: www.example.net

		user-agent: b
		disallow: /d

		host: example.com 
	`

	robots, _ := Parse(contents, url)

	if robots.Host() != "example.com" {
		t.Errorf("Expected host to be example.com")
	}
}

func TestRobotsTxt_parseEmptyAndInvalidDirectives(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent:
		user-agent:::: a::
		crawl-delay:
		crawl-delay:::: 0:
		host:
		host:: example.com
		sitemap:
		sitemap:: site:map.xml
		disallow:
		disallow::: /:
		allow:
		allow::: /:
	`

	_, err := Parse(contents, url)
	if err != nil {
		t.Errorf("Expected to not fail on invalid directives")
	}
}

func TestRobotsTxt_treatOnlyTheLastHostDirectiveAsValid(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: a
		crawl-delay: 1
		host: www.example.net

		user-agent: b
		disallow: /d

		host: example.net
		host: example.com 
	`

	robots, _ := Parse(contents, url)

	if robots.Host() != "example.com" {
		t.Errorf("Expected host to be example.com")
	}
}

func TestRobotsTxt_returnEmptyStringWhenThereIsNoHostDirective(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: a
		crawl-delay: 1

		user-agent: b
		disallow: /d
	`

	robots, _ := Parse(contents, url)

	if robots.Host() != "" {
		t.Errorf("Expected host to be empty")
	}
}

func TestRobotsTxt_fallbackToDefaultWhenUserAgentHasRulesOfItsOwn(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: *
		disallow: /test/
		crawl-delay: 1

		user-agent: b
		crawl-delay: 12

		user-agent: c
		user-agent: d
		crawl-delay: 10
	`

	robots, _ := Parse(contents, url)

	if robots.CrawlDelay("should-fall-back") != 1*time.Second {
		t.Errorf("Expected crawl delay for should-fall-back to be 1")
	}

	if robots.CrawlDelay("d") != 10*time.Second {
		t.Errorf("Expected crawl delay for d to be 10")
	}

	if robots.CrawlDelay("dd") != 1*time.Second {
		t.Errorf("Expected crawl delay for dd to be 1")
	}

	allowed, _ := robots.IsAllowed("should-fall-back", "http://www.example.com/test/")
	if allowed {
		t.Errorf("Expected /test/ to be disallowed for should-fall-back to")
	}
}

func TestRobotsTxt_shouldNotFallbackToDefaultWhenUserAgentHasRules(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: *
		disallow: /test/
		crawl-delay: 1

		user-agent: b
		allow:
	`

	robots, _ := Parse(contents, url)

	if robots.CrawlDelay("b") != 0*time.Second {
		t.Errorf("Expected crawl delay for b to be 0")
	}

	allowed, _ := robots.IsAllowed("b", "http://www.example.com/test/")
	if !allowed {
		t.Errorf("Expected /test/ to be allowed for b to")
	}
}

func TestRobotsTxt_ignoreVersionNumbersInTheUserAgentString(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		user-agent: *
		crawl-delay: 1

		user-agent: b
		crawl-delay: 12

		user-agent: c
		user-agent: d
		crawl-delay: 10
	`

	robots, _ := Parse(contents, url)

	if robots.CrawlDelay("should-fall-back/1.0.0") != time.Second {
		t.Errorf("Expected crawl delay for should-fall-back/1.0.0 to be 1")
	}

	if robots.CrawlDelay("d/12") != 10*time.Second {
		t.Errorf("Expected crawl delay for d/12 to be 10")
	}

	if robots.CrawlDelay("dd / 0-32-3") != 1*time.Second {
		t.Errorf("Expected crawl delay for dd / 0-32-3 to be 1")
	}

	if robots.CrawlDelay("b / 1.0") != 12*time.Second {
		t.Errorf("Expected crawl delay for b / 1.0 to be 12")
	}
}

func TestRobotsTxt_handleUrlEncodedAndUtf8PathsAndUrls(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent: *
		Disallow: /wiki:Article_wizard
		Disallow: /wiki%3AArticle_wizard
		Disallow: /اختبارات
		Disallow: /%E6%B5%8B%E8%AF%95
		Disallow: /考查
		Disallow: /%E0%A6%AA%E0%A6%B0%E0%A7%80%E0%A6%95%E0%A7%8D%E0%A6%B7%E0%A6%BE
	`

	allowed := []string{
		"http://www.example.com/fish",
		"http://www.example.com/اختبار",
	}

	disallowed := []string{
		"http://www.example.com/wiki:Article_wizard",
		"http://www.example.com/wiki%3AArticle_wizard",
		"http://www.example.com/اختبارات/test",
		"http://www.example.com/测试",
		"http://www.example.com/%E8%80%83%E6%9F%A5/test",
		"http://www.example.com/%E0%A6%AA%E0%A6%B0%E0%A7%80%E0%A6%95%E0%A7%8D%E0%A6%B7%E0%A6%BE",
	}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_invalidUrlEncodingsShouldBeTreatedAsUnencoded(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent: *
		Disallow: /%20%A/test
		Disallow: /%24%A/test$
		Disallow: /%B/*test%24
	`

	allowed := []string{
		"http://www.example.com/ %25A/test/test",
		"http://www.example.com/+%25A/test",
		"http://www.example.com/%20%25A/test",
		"http://www.example.com/*%25A/testing",
		"http://www.example.com/%25B/test$",
	}

	disallowed := []string{
		"http://www.example.com/%2520%25A/test",
		"http://www.example.com/%2524%25A/test",
		"http://www.example.com/%25B/test%2524",
	}

	testRobots(t, contents, url, allowed, disallowed)
}

func TestRobotsTxt_handleUrlEncodingsWithPatterns(t *testing.T) {
	url := "http://www.example.com/robots.txt"
	contents := `
		User-agent: *
		Disallow: /%20A/*test$
		Disallow: /%20B/*test%24
		Disallow: /%20C/test%24
		Disallow: /%20D/%2Atest$
	`

	allowed := []string{
		"http://www.example.com/ A/la/testing",
		"http://www.example.com/ B/la/test",
		"http://www.example.com/ C/test",
		"http://www.example.com/ D/la/test",
	}

	disallowed := []string{
		"http://www.example.com/ A/la/test",
		"http://www.example.com/ B/la/test$",
		"http://www.example.com/ B/la/test$test",
		"http://www.example.com/ C/test$",
		"http://www.example.com/ D/*test",
		"http://www.example.com/ D/%2Atest",
	}

	testRobots(t, contents, url, allowed, disallowed)
}
