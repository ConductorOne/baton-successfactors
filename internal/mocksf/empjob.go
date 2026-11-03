package mocksf

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type skipToken struct {
	StartRow int `json:"startRow"`
	EndRow   int `json:"endRow"`
}

func (s *server) handleEmpJob(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") || !s.validToken(strings.TrimPrefix(auth, "Bearer ")) {
		writeODataError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Synthetic: a valid bearer token is required.")
		return
	}

	q := r.URL.Query()
	start := 0
	if raw := q.Get("$skiptoken"); raw != "" {
		b, err := base64.StdEncoding.DecodeString(raw)
		var tok skipToken
		if err != nil || json.Unmarshal(b, &tok) != nil || tok.StartRow < 0 {
			writeODataError(w, http.StatusBadRequest, "BAD_REQUEST", "Synthetic: invalid $skiptoken.")
			return
		}
		start = tok.StartRow
	} else if q.Get("$format") != "json" || !strings.Contains(q.Get("$select"), "userId") ||
		!strings.Contains(q.Get("$expand"), "userNav") || !strings.Contains(q.Get("$expand"), "employmentNav") {
		// The first page carries the query that the connector builds.
		writeODataError(w, http.StatusBadRequest, "BAD_REQUEST", "Synthetic: $format=json, $select and $expand (userNav, employmentNav) are required.")
		return
	}

	page := start/s.cfg.PageSize + 1
	if s.cfg.FailPage == page && s.cfg.FailStatus != 0 {
		writeODataError(w, s.cfg.FailStatus, "INJECTED_FAILURE", "Synthetic: injected failure.")
		return
	}

	end := start + s.cfg.PageSize
	if end > len(s.cfg.Users) {
		end = len(s.cfg.Users)
	}
	if start > end {
		start = end
	}
	d := map[string]any{"results": s.cfg.Users[start:end]}
	if end < len(s.cfg.Users) {
		tok, _ := json.Marshal(skipToken{StartRow: end, EndRow: end + s.cfg.PageSize})
		next := url.Values{"$skiptoken": {base64.StdEncoding.EncodeToString(tok)}}
		d["__next"] = s.baseURL(r) + EmpJobPath + "?" + next.Encode()
	}
	if d["results"] == nil || len(s.cfg.Users[start:end]) == 0 {
		d["results"] = []json.RawMessage{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"d": d})
}
