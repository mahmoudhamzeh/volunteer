package httpserver

import (
	"net/http"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func (d Deps) myTickets(w http.ResponseWriter, r *http.Request) {
	if d.Tickets == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	items, err := d.Tickets.Mine(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) createTicket(w http.ResponseWriter, r *http.Request) {
	if d.Tickets == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	var in struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.Invalid("موضوع و متن تیکت را وارد کنید"))
		return
	}
	t, err := d.Tickets.Create(r.Context(), mustPrincipal(r).ID, in.Subject, in.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (d Deps) getMyTicket(w http.ResponseWriter, r *http.Request) {
	if d.Tickets == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	t, err := d.Tickets.GetMine(r.Context(), mustPrincipal(r).ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (d Deps) replyMyTicket(w http.ResponseWriter, r *http.Request) {
	if d.Tickets == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.Invalid("متن پیام را بنویسید"))
		return
	}
	t, err := d.Tickets.ReplyMine(r.Context(), mustPrincipal(r).ID, id, in.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (d Deps) adminTickets(w http.ResponseWriter, r *http.Request) {
	if d.Tickets == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	items, err := d.Tickets.List(r.Context(), domain.TicketStatus(r.URL.Query().Get("status")))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) adminTicket(w http.ResponseWriter, r *http.Request) {
	if d.Tickets == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	t, err := d.Tickets.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (d Deps) replyAdminTicket(w http.ResponseWriter, r *http.Request) {
	if d.Tickets == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.Invalid("متن پاسخ را بنویسید"))
		return
	}
	t, err := d.Tickets.ReplyAdmin(r.Context(), mustPrincipal(r).ID, id, in.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (d Deps) setTicketStatus(w http.ResponseWriter, r *http.Request) {
	if d.Tickets == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	t, err := d.Tickets.SetStatus(r.Context(), id, domain.TicketStatus(in.Status))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}
