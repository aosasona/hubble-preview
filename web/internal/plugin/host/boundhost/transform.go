package boundhost

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/go-shiori/go-readability"
	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/plugin/host/alloc"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
)

const DefaultReadabilityTimeout = 15 * time.Second

/*
Transform HTML to Markdown

Signature: fn(html: String) -> String (markdown)

Exported as: "transform_html_to_markdown"
*/
func (b *BoundHost) transformHtmlToMarkdown(
	ctx context.Context,
	m api.Module,
	offset, byteCount uint32,
) uint64 {
	logger := b.HostFnLogger(spec.PermTransformHtmlToMarkdown)

	buf, err := alloc.ReadBufferFromMemory(ctx, m, offset, byteCount)
	if err != nil {
		logger.error(err, "failed to read memory")
		return 0
	}

	markdown, err := htmltomarkdown.ConvertString(
		string(buf),
		converter.WithContext(ctx),
	)
	if err != nil {
		logger.error(err)
		return 0
	}

	encoded, err := alloc.WriteBufferToMemory(ctx, m, []byte(markdown))
	if err != nil {
		logger.error(err, "failed to write markdown to memory")
		return 0
	}

	return encoded
}

/*
Fetch the HTML content from a URL and transform it to Markdown.

Signature: fn(url: String) -> String (markdown)

Exported as: "transform_url_to_markdown"
*/
func (b *BoundHost) transformUrlToMarkdown(
	ctx context.Context,
	m api.Module,
	offset, byteCount uint32,
) uint64 {
	logger := b.HostFnLogger(spec.PermTransformUrlToMarkdown)

	buf, err := alloc.ReadBufferFromMemory(ctx, m, offset, byteCount)
	if err != nil {
		logger.error(err, "failed to read memory")
		return 0
	}

	link := string(buf)
	article, err := readability.FromURL(link, DefaultReadabilityTimeout, func(req *http.Request) {
		req.Header.Set(
			"User-Agent",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
		)
	})
	if err != nil {
		logger.error(err, "failed to load article")
		return 0
	}

	parsedUrl, _ := url.Parse(link)
	baseUrl := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)

	markdown, err := htmltomarkdown.ConvertString(
		article.Content,
		converter.WithContext(ctx),
		converter.WithDomain(baseUrl),
	)
	if err != nil {
		logger.error(err, "failed to convert article to markdown")
		return 0
	}

	encoded, err := alloc.WriteBufferToMemory(ctx, m, []byte(markdown))
	if err != nil {
		logger.error(err, "failed to write markdown to memory")
		return 0
	}

	return encoded
}
