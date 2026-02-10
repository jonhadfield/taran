package email

import (
	"strings"

	"golang.org/x/net/html"
)

// HTMLToText converts HTML content to readable plain text.
func HTMLToText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var b strings.Builder
	extractText(doc, &b)
	return strings.TrimSpace(b.String())
}

func extractText(n *html.Node, b *strings.Builder) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "script", "style", "head":
			return
		case "br":
			b.WriteString("\n")
		case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li", "tr":
			if b.Len() > 0 {
				b.WriteString("\n\n")
			}
		}
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			if b.Len() > 0 && !strings.HasSuffix(b.String(), "\n") {
				b.WriteString(" ")
			}
			b.WriteString(text)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, b)
	}
}
