package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const noGroupHTML = `<!DOCTYPE html><html><head><title>Steam Community :: Error</title></head>
<body><h3>No group could be retrieved for the given URL.</h3></body></html>`

const validXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<memberList>
  <groupID64>103582791460000000</groupID64>
  <groupDetails>
    <groupName>Test Game</groupName>
    <memberCount>4242</memberCount>
  </groupDetails>
</memberList>`

func TestFetchMembersList_NoGroupReturnsSentinel(t *testing.T) {
	srv := httptest.NewServer(handlerString(noGroupHTML))
	defer srv.Close()

	_, err := fetchMembersList(srv.URL)
	if !errors.Is(err, errNoGroup) {
		t.Fatalf("expected errNoGroup, got %v", err)
	}
}

func TestFetchMembersList_ValidXML(t *testing.T) {
	srv := httptest.NewServer(handlerString(validXML))
	defer srv.Close()

	details, err := fetchMembersList(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.MemberCount != 4242 {
		t.Errorf("MemberCount = %d, want 4242", details.MemberCount)
	}
	if details.GroupName != "Test Game" {
		t.Errorf("GroupName = %q, want %q", details.GroupName, "Test Game")
	}
}

func TestGetMembersList_ShortCircuitsOnNoGroup(t *testing.T) {
	srv := httptest.NewServer(handlerString(noGroupHTML))
	defer srv.Close()

	// getMembersList builds the URL from membersListUrl. Override it to point
	// at the test server so we can assert no-retry behavior end-to-end.
	orig := membersListUrl
	membersListUrl = srv.URL + "/?app=%d"
	defer func() { membersListUrl = orig }()

	start := time.Now()
	_, err := getMembersList(1536500)
	elapsed := time.Since(start)

	if !errors.Is(err, errNoGroup) {
		t.Fatalf("expected errNoGroup, got %v", err)
	}
	// First quick retry waits 5s; if we short-circuited correctly we should be
	// well under that.
	if elapsed > 2*time.Second {
		t.Errorf("getMembersList took %v; expected to short-circuit immediately", elapsed)
	}
}

func handlerString(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = strings.NewReader(body).WriteTo(w)
	}
}
