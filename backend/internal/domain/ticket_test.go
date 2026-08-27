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
