package domain

import "testing"

func TestValidTicketSubject(t *testing.T) {
	if !ValidTicketSubject("آموزش") {
		t.Fatal("expected catalog subject to be valid")
	}
	if ValidTicketSubject("موضوع آزاد") {
		t.Fatal("free-text subject must be rejected")
	}
	if ValidTicketSubject("") {
		t.Fatal("empty subject must be rejected")
	}
}

func TestNormalizeTicketQuery(t *testing.T) {
	text, digits := NormalizeTicketQuery(" #۱۲۳ ")
	if text != "123" || digits != "123" {
		t.Fatalf("got text=%q digits=%q", text, digits)
	}
	text, digits = NormalizeTicketQuery("سارا محمدی")
	if text != "سارا محمدی" || digits != "" {
		t.Fatalf("got text=%q digits=%q", text, digits)
	}
	text, digits = NormalizeTicketQuery("۰۹۱۲۱۲۳۴۵۶۷")
	if text != "09121234567" || digits != "09121234567" {
		t.Fatalf("got text=%q digits=%q", text, digits)
	}
}
