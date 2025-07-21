package ograph

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

type Extractor struct {
	// The URL of the page.
	link *url.URL

	// The domain of the page.
	domain *url.URL

	// The document of the page.
	document *goquery.Document
}

func NewExtractor(link *url.URL, document *goquery.Document) (*Extractor, error) {
	domain, err := url.Parse(fmt.Sprintf("%s://%s", link.Scheme, link.Hostname()))
	if err != nil {
		return nil, err
	}

	return &Extractor{
		link:     link,
		domain:   domain,
		document: document,
	}, nil
}

func (e *Extractor) Extract() (*Metadata, error) {
	var err error
	meta := &Metadata{Domain: e.domain.String(), Link: e.link.String()}

	meta.Title = e.getTitle()
	meta.Description = e.getDescription()

	meta.Favicon, err = e.getFavicon()
	if err != nil {
		return nil, err
	}

	meta.Author = e.getAuthor()
	meta.Thumbnail = e.getThumbnail()
	meta.SiteType = e.getSiteType()

	return meta, nil
}

// getMeta returns the content of the meta tag with the given name.
func (e *Extractor) getMeta(name string) string {
	// Find the meta tag by name.
	meta := e.document.Find(fmt.Sprintf("meta[name='%s']", name))
	if meta.Length() == 0 {
		// If the meta tag is not found, try to find it by property.
		meta = e.document.Find(fmt.Sprintf("meta[property='%s']", name))
	}

	return meta.AttrOr("content", "")
}

func (e *Extractor) getTitle() string {
	var title string

	title = e.getMeta("og:title")
	if strings.TrimSpace(title) == "" {
		title = e.document.Find("title").Text()
	}

	return strings.TrimSpace(title)
}

func (e *Extractor) getDescription() string {
	var description string

	description = e.getMeta("og:description")
	if strings.TrimSpace(description) == "" {
		description = e.getMeta("description")
	}

	return description
}

func (e *Extractor) getFavicon() (string, error) {
	var favicon string

	// Iterate through the probable selectors of the favicon and select the first one that exists.
	selectors := []string{
		"icon",                         // Favicon
		"mask-icon",                    // Mask icon
		"shortcut icon",                // Shortcut icon
		"apple-touch-icon",             // Apple touch icon
		"apple-touch-icon-precomposed", // Apple touch icon (precomposed)
	}

	for _, selector := range selectors {
		e.document.Find(fmt.Sprintf("link[rel='%s']", selector)).
			Each(func(i int, selection *goquery.Selection) {
				icon := selection.AttrOr("href", "")
				if strings.TrimSpace(icon) != "" {
					favicon = icon
					return
				}
			})
	}

	return e.resolveUrlWithDomain(favicon), nil
}

func (e *Extractor) getAuthor() string {
	return e.getMeta("author")
}

func (e *Extractor) getThumbnail() string {
	var thumbnailLink string

	selectors := []string{
		"og:image",                // Open Graph image
		"twitter:image",           // Twitter image
		"image",                   // Image
		"msapplication-TileImage", // Microsoft application tile image
		"thumbnail",               // Thumbnail
	}

	for _, selector := range selectors {
		link := e.getMeta(selector)
		if strings.TrimSpace(link) != "" {
			thumbnailLink = link
			break
		}
	}

	// Check for `image_src` as a fallback.
	if strings.TrimSpace(thumbnailLink) == "" {
		thumbnailLink = e.document.Find(fmt.Sprintf("link[rel='%s']", "image_src")).
			AttrOr("href", "")
	}

	thumbnailUrl, err := url.Parse(thumbnailLink)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse thumbnail URL")
		return ""
	}

	return e.resolveUrlWithDomain(thumbnailUrl.String())
}

func (e *Extractor) getSiteType() string {
	var siteType string

	siteType = e.getMeta("og:type")
	if strings.TrimSpace(siteType) == "" {
		siteType = "website"
	}

	return siteType
}

func (e *Extractor) resolveUrlWithDomain(href string) string {
	if strings.TrimSpace(href) == "" {
		return ""
	}

	// If it starts with //, resolve it with the domain.
	if strings.HasPrefix(href, "//") {
		return fmt.Sprintf("%s:%s", e.link.Scheme, href)
	}

	// If it starts with /, i.e. it is a relative URL, resolve it with the domain.
	if strings.HasPrefix(href, "/") {
		return fmt.Sprintf("%s://%s%s", e.link.Scheme, e.link.Hostname(), href)
	}

	return href
}
