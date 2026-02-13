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

// HTMLToMarkdown converts HTML email content to a simplified markdown-like
// representation that preserves headings, links, lists, and emphasis while
// stripping layout/style noise. This gives the LLM structural context that
// plain text extraction loses.
func HTMLToMarkdown(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return HTMLToText(htmlContent)
	}

	var b strings.Builder
	htmlToMd(doc, &b)

	// Collapse excessive blank lines
	result := strings.TrimSpace(b.String())
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return result
}

func htmlToMd(n *html.Node, b *strings.Builder) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "script", "style", "head":
			return
		case "img":
			// Skip images (tracking pixels, decorative)
			return
		}
	}

	// Opening tags
	if n.Type == html.ElementNode {
		switch n.Data {
		case "h1":
			b.WriteString("\n\n# ")
		case "h2":
			b.WriteString("\n\n## ")
		case "h3":
			b.WriteString("\n\n### ")
		case "h4", "h5", "h6":
			b.WriteString("\n\n#### ")
		case "p", "div", "section", "article":
			b.WriteString("\n\n")
		case "br":
			b.WriteString("\n")
		case "li":
			b.WriteString("\n- ")
		case "ul", "ol":
			b.WriteString("\n")
		case "tr":
			b.WriteString("\n")
		case "td", "th":
			b.WriteString(" | ")
		case "strong", "b":
			b.WriteString("**")
		case "em", "i":
			b.WriteString("*")
		case "a":
			// Will be handled after children
		case "blockquote":
			b.WriteString("\n\n> ")
		case "hr":
			b.WriteString("\n\n---\n\n")
		}
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			// Add space separator if needed
			if b.Len() > 0 {
				last := b.String()[b.Len()-1]
				if last != '\n' && last != ' ' && last != '*' && last != '#' {
					b.WriteString(" ")
				}
			}
			b.WriteString(text)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		htmlToMd(c, b)
	}

	// Closing tags
	if n.Type == html.ElementNode {
		switch n.Data {
		case "strong", "b":
			b.WriteString("**")
		case "em", "i":
			b.WriteString("*")
		case "a":
			href := getAttr(n, "href")
			if href != "" && !strings.HasPrefix(href, "#") && !strings.Contains(href, "unsubscribe") {
				b.WriteString(" (")
				b.WriteString(href)
				b.WriteString(")")
			}
		}
	}
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
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
