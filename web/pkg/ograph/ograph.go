// ograph is a package that provides functionality to parse Open Graph metadata from a URL.
package ograph

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.trulyao.dev/seer"
)

type (
	Metadata struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Favicon     string `json:"favicon"`
		Author      string `json:"author"`
		Thumbnail   string `json:"thumbnail"`
		SiteType    string `json:"site_type"`

		Domain string `json:"domain"`
		Link   string `json:"link"`
	}

	Options struct {
		// Headers to send with the request.
		Headers map[string]string
	}
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("stopped after 10 redirects")
			}

			return nil
		},
	}

	defaultHeaders = map[string]string{
		// Use googlebot as the user agent to get the most accurate results.
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
	}
)

func Parse(url string, options ...Options) (*Metadata, error) {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	metadata := new(Metadata)

	// Parse the URL.
	u, err := UrlFromString(url)
	if err != nil {
		return metadata, err
	}

	doc, err := RequestDocument(u, opts)
	if err != nil {
		return metadata, err
	}

	extractor, err := NewExtractor(u, doc)
	if err != nil {
		return metadata, err
	}

	return extractor.Extract()
}

func RequestDocument(u *url.URL, opts Options) (*goquery.Document, error) {
	request, err := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	// Set the headers.
	if len(opts.Headers) > 0 {
		for k, v := range opts.Headers {
			request.Header.Set(k, v)
		}
	}

	for k, v := range defaultHeaders {
		request.Header.Set(k, v)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, seer.Wrap("http_client_request", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, seer.Wrap(
			"http_client_request",
			fmt.Errorf("unexpected status code: %d", response.StatusCode),
			"unable to load the page",
		)
	}
	defer response.Body.Close() // Close body when done

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, seer.Wrap("goquery_new_document", err)
	}

	return doc, nil
}

// UrlFromString parses a URL string and returns a URL object.
func UrlFromString(link string) (*url.URL, error) {
	var (
		u   *url.URL
		err error
	)

	if link == "" {
		return &url.URL{}, errors.New("empty URL")
	}

	if u, err = url.Parse(link); err != nil {
		return &url.URL{}, err
	}

	if u.Scheme == "" {
		return &url.URL{}, errors.New("missing URL scheme")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return &url.URL{}, errors.New("unsupported URL scheme, only http and https are supported")
	}

	if strings.Count(u.Hostname(), ".") < 1 {
		return &url.URL{}, errors.New("invalid URL host, must be in the format 'example.com'")
	}

	return u, nil
}
