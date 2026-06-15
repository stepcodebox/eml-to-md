package converter

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

type Attachment struct {
	Filename    string
	ContentType string
	Size        int
}

type messageContent struct {
	Header      mail.Header
	PlainText   string
	HTML        string
	Attachments []Attachment
}

func Convert(r io.Reader) (string, error) {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return "", err
	}

	content := messageContent{Header: msg.Header}
	if err := collectParts(msg.Header, msg.Body, &content); err != nil {
		return "", err
	}

	body := strings.TrimSpace(content.PlainText)
	if strings.TrimSpace(content.HTML) != "" {
		body, err = htmlToMarkdown(content.HTML)
		if err != nil {
			return "", err
		}
	}
	if strings.TrimSpace(body) == "" {
		body = "*(empty body)*"
	}

	return renderMarkdown(content, body), nil
}

func collectParts(header mail.Header, body io.Reader, content *messageContent) error {
	mediaType, params, _ := mime.ParseMediaType(header.Get("Content-Type"))
	if mediaType == "" {
		mediaType = "text/plain"
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return nil
		}
		reader := multipart.NewReader(body, boundary)
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if err := collectPart(part, content); err != nil {
				return err
			}
		}
		return nil
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	addBodyPart(mediaType, string(data), content)
	return nil
}

func collectPart(part *multipart.Part, content *messageContent) error {
	disposition, dispositionParams, _ := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
	mediaType, params, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
	if mediaType == "" {
		mediaType = "text/plain"
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		return collectParts(mail.Header(part.Header), part, content)
	}

	data, err := io.ReadAll(part)
	if err != nil {
		return err
	}

	filename := part.FileName()
	if filename == "" {
		filename = dispositionParams["filename"]
	}
	if filename == "" {
		filename = params["name"]
	}

	if disposition == "attachment" || filename != "" {
		if filename == "" {
			filename = "unnamed"
		}
		content.Attachments = append(content.Attachments, Attachment{
			Filename:    filename,
			ContentType: mediaType,
			Size:        len(data),
		})
		return nil
	}

	addBodyPart(mediaType, string(data), content)
	return nil
}

func addBodyPart(mediaType string, data string, content *messageContent) {
	switch strings.ToLower(mediaType) {
	case "text/html":
		if content.HTML == "" {
			content.HTML = data
		}
	case "text/plain":
		if content.PlainText == "" {
			content.PlainText = data
		}
	}
}

func htmlToMarkdown(html string) (string, error) {
	out, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func renderMarkdown(content messageContent, body string) string {
	var b strings.Builder
	subject := header(content.Header, "Subject", "(no subject)")

	fmt.Fprintf(&b, "# %s\n\n", subject)
	writeHeader(&b, "From", content.Header.Get("From"))
	writeHeader(&b, "To", content.Header.Get("To"))
	writeHeader(&b, "CC", content.Header.Get("Cc"))
	writeHeader(&b, "Reply-To", content.Header.Get("Reply-To"))
	writeHeader(&b, "Date", content.Header.Get("Date"))

	b.WriteString("\n---\n\n")
	b.WriteString(body)
	b.WriteString("\n")

	if len(content.Attachments) > 0 {
		b.WriteString("\n---\n\n## Attachments\n\n")
		for _, attachment := range content.Attachments {
			fmt.Fprintf(
				&b,
				"- `%s` (%s, %s)\n",
				attachment.Filename,
				formatSize(attachment.Size),
				attachment.ContentType,
			)
		}
	}

	return b.String()
}

func writeHeader(b *strings.Builder, label string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	fmt.Fprintf(b, "- **%s:** %s\n", label, strings.TrimSpace(value))
}

func header(h mail.Header, key string, fallback string) string {
	value := strings.TrimSpace(h.Get(key))
	if value == "" {
		return fallback
	}
	decoder := mime.WordDecoder{}
	decoded, err := decoder.DecodeHeader(value)
	if err != nil {
		return value
	}
	return decoded
}

func formatSize(size int) string {
	switch {
	case size >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func ConvertBytes(data []byte) (string, error) {
	return Convert(bytes.NewReader(data))
}
