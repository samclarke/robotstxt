// Package robotstxt parses robots.txt files
package robotstxt

// Aims to follow the Google specification, see:
// https://developers.google.com/search/reference/robots_txt
// for more information.

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

type rule struct {
	isPattern bool
	isAllowed bool
	path      string
	pattern   *regexp.Regexp
}

type userAgentRules struct {
	rules      []*rule
	crawlDelay float32
}

// RobotsTxt represents a parsed robots.txt file
type RobotsTxt struct {
	url            *url.URL
	userAgentRules map[string]*userAgentRules
	sitemaps       []string
	host           string
}

// InvalidHostError is the error when a URL is tested with IsAllowed that
// is not valid for this robots.txt file
type InvalidHostError struct{}

func (e InvalidHostError) Error() string {
	return "URL is not valid for this robots.txt file"
}

func parseAndNormalizeURL(urlStr string) (u *url.URL, err error) {
	u, err = url.Parse(urlStr)
	if err != nil {
		return
	}

	u.Host, err = idna.ToASCII(u.Host)
	if err != nil {
		return
	}

	return
}

func replaceSuffix(s, suffix, replacement string) string {
	if strings.HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)] + replacement
	}
	return s
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

func (r *userAgentRules) isAllowed(userAgent string, path string) bool {
	var result = true
	var resultPathLength = 0

	for _, rule := range r.rules {
		if rule.isPattern {
			// The first matching pattern takes precedence
			if rule.pattern.MatchString(path) {
				return rule.isAllowed
			}
		} else {
			// The longest matching path precedence
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
		url:            u,
		userAgentRules: make(map[string]*userAgentRules),
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
	agentRules, ok := r.userAgentRules[userAgent]
	if !ok {
		agentRules = &userAgentRules{}
		r.userAgentRules[userAgent] = agentRules
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

		agentRules.rules = append(agentRules.rules, &rule{
			isPattern: true,
			pattern:   regexPattern,
			isAllowed: isAllowed,
		})
	} else {
		agentRules.rules = append(agentRules.rules, &rule{
			isPattern: false,
			path:      path,
			isAllowed: isAllowed,
		})
	}

	return nil
}

func (r *RobotsTxt) addCrawlDelay(userAgent string, crawlDelay string) (err error) {
	agentRules, ok := r.userAgentRules[userAgent]
	if !ok {
		agentRules = &userAgentRules{}
		r.userAgentRules[userAgent] = agentRules
	}

	if delay, err := strconv.ParseFloat(crawlDelay, 32); err == nil {
		agentRules.crawlDelay = float32(delay)
	}

	return
}

// Host is the preferred hosts from the robots.txt file if there is one
func (r *RobotsTxt) Host() string {
	return r.host
}

// CrawlDelay returns the crawl delay for the specified
// user agent or 0 if there is none
func (r *RobotsTxt) CrawlDelay(userAgent string) float32 {
	userAgent = normaliseUserAgent(userAgent)

	if agentRules, ok := r.userAgentRules[userAgent]; ok {
		return agentRules.crawlDelay
	}

	if agentRules, ok := r.userAgentRules["*"]; ok {
		return agentRules.crawlDelay
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

	if rules, ok := r.userAgentRules[normaliseUserAgent(userAgent)]; ok {
		result = rules.isAllowed(userAgent, u.Path)
	} else if rules, ok := r.userAgentRules["*"]; ok {
		result = rules.isAllowed(userAgent, u.Path)
	}

	return
}
