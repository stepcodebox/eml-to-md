package converter

import (
	"strings"
	"testing"
)

func TestConvertPlainText(t *testing.T) {
	md, err := ConvertBytes([]byte(plainTextEML))
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, md, "# Plain text test")
	assertContains(t, md, "- **From:** alice@example.com")
	assertContains(t, md, "This is a plain text email.")
}

func TestConvertHTMLPreferred(t *testing.T) {
	md, err := ConvertBytes([]byte(htmlEML))
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, md, "**HTML**")
	assertContains(t, md, "[a link](https://example.com)")
	if strings.Contains(md, "Plain text fallback") {
		t.Fatalf("expected HTML body to be preferred, got:\n%s", md)
	}
}

func TestConvertAttachments(t *testing.T) {
	md, err := ConvertBytes([]byte(attachmentEML))
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, md, "## Attachments")
	assertContains(t, md, "`notes.txt`")
	assertContains(t, md, "text/plain")
}

func TestMissingBody(t *testing.T) {
	md, err := ConvertBytes([]byte("From: sender@example.com\r\n\r\n"))
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, md, "# (no subject)")
	assertContains(t, md, "*(empty body)*")
}

func assertContains(t *testing.T, haystack string, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected %q in:\n%s", needle, haystack)
	}
}

const plainTextEML = "From: alice@example.com\r\n" +
	"To: bob@example.com\r\n" +
	"Subject: Plain text test\r\n" +
	"Date: Mon, 10 Mar 2026 10:00:00 +0000\r\n" +
	"Content-Type: text/plain; charset=utf-8\r\n" +
	"\r\n" +
	"This is a plain text email.\r\n\r\nBest,\r\nAlice\r\n"

const htmlEML = "From: alice@example.com\r\n" +
	"To: bob@example.com\r\n" +
	"Subject: HTML test\r\n" +
	"Content-Type: multipart/alternative; boundary=alt\r\n" +
	"\r\n" +
	"--alt\r\n" +
	"Content-Type: text/plain; charset=utf-8\r\n" +
	"\r\n" +
	"Plain text fallback\r\n" +
	"--alt\r\n" +
	"Content-Type: text/html; charset=utf-8\r\n" +
	"\r\n" +
	"<p><strong>HTML</strong> body with <a href=\"https://example.com\">a link</a>.</p>\r\n" +
	"--alt--\r\n"

const attachmentEML = "From: alice@example.com\r\n" +
	"To: bob@example.com\r\n" +
	"Subject: Attachment test\r\n" +
	"Content-Type: multipart/mixed; boundary=mix\r\n" +
	"\r\n" +
	"--mix\r\n" +
	"Content-Type: text/plain; charset=utf-8\r\n" +
	"\r\n" +
	"See attached.\r\n" +
	"--mix\r\n" +
	"Content-Type: text/plain; name=notes.txt\r\n" +
	"Content-Disposition: attachment; filename=notes.txt\r\n" +
	"\r\n" +
	"notes\r\n" +
	"--mix--\r\n"
