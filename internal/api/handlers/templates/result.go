package templates

import (
	"fmt"
	"html/template"
)

type TemplateDataResult struct {
	Title   string
	Content template.HTML
}

// htmlTemplate is the template for wrapping the content
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        /* Base styles */
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #24292e;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
        }

        /* Markdown content styles */
        .markdown-body h1, .markdown-body h2, .markdown-body h3, 
        .markdown-body h4, .markdown-body h5, .markdown-body h6 {
            margin-top: 24px;
            margin-bottom: 16px;
            font-weight: 600;
            line-height: 1.25;
        }

        .markdown-body h1 { font-size: 2em; }
        .markdown-body h2 { font-size: 1.5em; }
        .markdown-body h3 { font-size: 1.25em; }

        .markdown-body p {
            margin-bottom: 16px;
        }

        .markdown-body code {
            padding: 0.2em 0.4em;
            margin: 0;
            font-size: 85%;
            background-color: rgba(27, 31, 35, 0.05);
            border-radius: 3px;
            font-family: SFMono-Regular, Consolas, 'Liberation Mono', Menlo, monospace;
        }

        .markdown-body pre {
            padding: 16px;
            overflow: auto;
            font-size: 85%;
            line-height: 1.45;
            background-color: #f6f8fa;
            border-radius: 3px;
        }

        .markdown-body pre code {
            display: block;
            padding: 0;
            margin: 0;
            overflow: visible;
            line-height: inherit;
            word-wrap: normal;
            background-color: transparent;
            border: 0;
        }

        .markdown-body blockquote {
            padding: 0 1em;
            color: #6a737d;
            border-left: 0.25em solid #dfe2e5;
            margin: 0 0 16px 0;
        }

        .markdown-body ul, .markdown-body ol {
            padding-left: 2em;
            margin-bottom: 16px;
        }

        .markdown-body table {
            border-spacing: 0;
            border-collapse: collapse;
            margin-bottom: 16px;
            width: 100%;
        }

        .markdown-body table th,
        .markdown-body table td {
            padding: 6px 13px;
            border: 1px solid #dfe2e5;
        }

        .markdown-body table tr {
            background-color: #fff;
            border-top: 1px solid #c6cbd1;
        }

        .markdown-body table tr:nth-child(2n) {
            background-color: #f6f8fa;
        }

        /* Status badges */
        .status {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 3px;
            font-weight: 600;
            font-size: 12px;
            margin-bottom: 16px;
        }
        .status-success { background-color: #28a745; color: white; }
        .status-error { background-color: #dc3545; color: white; }
        .status-warning { background-color: #ffc107; color: black; }
        .status-pending { background-color: #6c757d; color: white; }
    </style>
</head>
<body>
    <div class="markdown-body">
        {{.Content}}
    </div>
</body>
</html>
`

func GetResultTemplate() *template.Template {
	tmpl, err := template.New("results").Parse(htmlTemplate)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse HTML template: %v", err))
	}
	return tmpl
}
