// Package robotstxt parses robots.txt files
//
// Aims to follow the Google robots.txt specification, see:
// https://developers.google.com/search/reference/robots_txt
// for more information.
package robotstxt

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/idna"
)

type rule struct {
	isAllowed bool
	path      string
	pattern   *regexp.Regexp
}

type group struct {
	rules      []*rule
	crawlDelay time.Duration
}

// RobotsTxt represents a parsed robots.txt file
type RobotsTxt struct {
	url      *url.URL
	groups   map[string]*group
	sitemaps []string
	host     string
}

// InvalidHostError is the error when a URL is tested with IsAllowed that
// is not valid for this robots.txt file
type InvalidHostError struct{}

func (e InvalidHostError) Error() string {
	return "URL is not valid for this robots.txt file"
}

func parseAndNormalizeURL(urlStr string) (u *url.URL, err error) {
	u, err = url.Parse(urlStr)
	if err == nil {
		u.Host, err = idna.ToASCII(u.Host)
	}

	return
}

func replaceSuffix(str, suffix, replacement string) string {
	if strings.HasSuffix(str, suffix) {
		return str[:len(str)-len(suffix)] + replacement
	}

	return str
}

func isPattern(path string) bool {
	return strings.IndexRune(path, '*') > -1 || strings.HasSuffix(path, "$")
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.Replace(pattern, "\\*", "(?:.*)", -1)

	pattern = replaceSuffix(pattern, "\\$", "$")
	pattern = replaceSuffix(pattern, "%24", "\\$")
	pattern = replaceSuffix(pattern, "%2524", "%24")

	pattern = strings.Replace(pattern, "%2A", "\\*", -1)

	return regexp.Compile(pattern)
}

func normaliseUserAgent(userAgent string) string {
	index := strings.IndexRune(userAgent, '/')
	if index > -1 {
		userAgent = userAgent[:index]
	}

	return strings.ToLower(strings.TrimSpace(userAgent))
}

func (r *group) isAllowed(userAgent string, path string) bool {
	var result = true
	var resultPathLength = 0

	for _, rule := range r.rules {
		if rule.pattern != nil {
			// The first matching pattern takes precedence
			if rule.pattern.MatchString(path) {
				return rule.isAllowed
			}
		} else {
			// The longest matching path takes precedence
			if resultPathLength > len(rule.path) {
				continue
			}

			if strings.HasPrefix(path, rule.path) {
				result = rule.isAllowed
				resultPathLength = len(rule.path)
			}
		}
	}

	return result
}

// Parse parses the contents or a robots.txt file and returns a
// RobotsTxt struct that can be used to check if URLs can be crawled
// or extract crawl delays, sitemaps or the preferred host name
func Parse(contents string, urlStr string) (robotsTxt *RobotsTxt, err error) {
	u, err := parseAndNormalizeURL(urlStr)
	if err != nil {
		return
	}

	robotsTxt = &RobotsTxt{
		url:    u,
		groups: make(map[string]*group),
	}

	var userAgents []string
	isNoneUserAgentState := false

	lines := strings.Split(contents, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) > 1 {
			rule, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

			switch strings.ToLower(rule) {
			case "user-agent":
				if isNoneUserAgentState {
					userAgents = nil
				}
				userAgents = append(userAgents, normaliseUserAgent(val))
				break
			case "allow":
				for _, ua := range userAgents {
					robotsTxt.addPathRule(ua, val, true)
				}
				break
			case "disallow":
				for _, ua := range userAgents {
					robotsTxt.addPathRule(ua, val, false)
				}
				break
			case "crawl-delay":
				for _, ua := range userAgents {
					robotsTxt.addCrawlDelay(ua, val)
				}
				break
			case "sitemap":
				if val != "" {
					robotsTxt.sitemaps = append(robotsTxt.sitemaps, val)
				}
				break
			case "host":
				if val != "" {
					robotsTxt.host = val
				}
				break
			}

			isNoneUserAgentState = strings.ToLower(rule) != "user-agent"
		}
	}

	return
}

func (r *RobotsTxt) addPathRule(userAgent string, path string, isAllowed bool) error {
	g, ok := r.groups[userAgent]
	if !ok {
		g = &group{}
		r.groups[userAgent] = g
	}

	isPattern := isPattern(path)
	if isPattern {
		path = replaceSuffix(path, "%24", "%2524")
	}

	// Keep * escaped
	path = strings.Replace(path, "%2A", "%252A", -1)
	if unescapedPath, err := url.PathUnescape(path); err == nil {
		path = unescapedPath
	} else {
		path = strings.Replace(path, "%252A", "%2A", -1)
	}

	if isPattern {
		regexPattern, err := compilePattern(path)
		if err != nil {
			return err
		}

		g.rules = append(g.rules, &rule{
			pattern:   regexPattern,
			isAllowed: isAllowed,
		})
	} else {
		g.rules = append(g.rules, &rule{
			path:      path,
			isAllowed: isAllowed,
		})
	}

	return nil
}

func (r *RobotsTxt) addCrawlDelay(userAgent string, crawlDelay string) (err error) {
	g, ok := r.groups[userAgent]
	if !ok {
		g = &group{}
		r.groups[userAgent] = g
	}

	if delay, err := strconv.ParseFloat(crawlDelay, 64); err == nil {
		g.crawlDelay = time.Duration(delay * float64(time.Second))
	}

	return
}

// Host is the preferred hosts from the robots.txt file if there is one
func (r *RobotsTxt) Host() string {
	return r.host
}

// CrawlDelay returns the crawl delay for the specified
// user agent or 0 if there is none
func (r *RobotsTxt) CrawlDelay(userAgent string) time.Duration {
	if group, ok := r.groups[normaliseUserAgent(userAgent)]; ok {
		return group.crawlDelay
	}

	if group, ok := r.groups["*"]; ok {
		return group.crawlDelay
	}

	return 0
}

// Sitemaps returns a list of sitemaps from the robots.txt file if any
func (r *RobotsTxt) Sitemaps() []string {
	return r.sitemaps
}

// IsAllowed checks if the specified path is allowed by the robots.txt file
func (r *RobotsTxt) IsAllowed(userAgent string, urlStr string) (result bool, err error) {
	u, err := parseAndNormalizeURL(urlStr)
	if err != nil {
		return
	}

	if u.Scheme != r.url.Scheme || u.Host != r.url.Host {
		err = &InvalidHostError{}
		return
	}

	result = true

	if group, ok := r.groups[normaliseUserAgent(userAgent)]; ok {
		result = group.isAllowed(userAgent, u.Path)
	} else if group, ok := r.groups["*"]; ok {
		result = group.isAllowed(userAgent, u.Path)
	}

	return
}
