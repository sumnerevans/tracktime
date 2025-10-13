package report

import (
	"bytes"
	"fmt"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"go.mau.fi/util/exerrors"
)

// htmlTemplate wraps the markdown-generated HTML with proper HTML structure and CSS
const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body {
            background-color: white;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 20px;
        }

        .content {
            max-width: 900px;
            margin: 0 auto;
        }

        h1, h2 {
            color: #2c3e50;
        }

        h1 {
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }

        h2 {
            margin-top: 30px;
            border-bottom: 1px solid #bdc3c7;
            padding-bottom: 8px;
        }

        table {
            border-collapse: collapse;
            width: 100%%;
            margin: 20px 0;
        }

        th {
            background-color: #3498db;
            color: white;
            font-weight: bold;
            padding: 12px 8px;
            text-align: left;
            border: 1px solid #2980b9;
        }

        td {
            padding: 10px 8px;
            border: 1px solid #ddd;
        }

        /* Right-align numeric columns (Hours, Rate, Total) */
        td:nth-child(2), td:nth-child(3), td:nth-child(4) {
            text-align: right;
        }

        th:nth-child(2), th:nth-child(3), th:nth-child(4) {
            text-align: right;
        }

        tbody tr:hover {
            background-color: #f5f5f5;
        }

        /* Alternate row colors for better readability */
        tbody tr:nth-child(even) {
            background-color: #f9f9f9;
        }

        /* Bold text styling */
        strong {
            font-weight: bold;
            color: #2c3e50;
        }

        /* List styling for customer addresses */
        ul {
            list-style-type: none;
            padding-left: 0;
        }

        li {
            padding: 2px 0;
        }

        /* Footer note styling */
        p {
            color: #7f8c8d;
            font-style: italic;
        }
    </style>
</head>
<body>
    <div class="content">
%s
    </div>
</body>
</html>`

// GenerateHTMLReport generates an HTML report by converting markdown to HTML
func (r *Report) GenerateHTMLReport(w io.Writer) {
	// Generate markdown to buffer
	var markdownBuf bytes.Buffer
	r.GenerateMarkdownReport(&markdownBuf)

	// Configure goldmark with table extension
	md := goldmark.New(goldmark.WithExtensions(extension.Table))

	// Convert markdown to HTML
	var htmlBody bytes.Buffer
	exerrors.PanicIfNotNil(md.Convert(markdownBuf.Bytes(), &htmlBody))

	// Write wrapped HTML to writer
	exerrors.Must(fmt.Fprintf(w, htmlTemplate, r.headerText(), htmlBody.String()))
}
