package admin_test

import (
	"rustdesk-api-server-pro/app/controller/admin"
	"testing"
)

func TestParsePeer(t *testing.T) {
	cases := []struct {
		peer     string
		wantId   string
		wantName string
	}{
		{`["222222222","Controller PC"]`, "222222222", "Controller PC"},
		{`["222222222"]`, "222222222", ""},
		{"222222222", "222222222", ""},
		{"", "", ""},
	}

	for _, c := range cases {
		peerId, peerName := admin.ParsePeer(c.peer)
		if peerId != c.wantId || peerName != c.wantName {
			t.Fatalf("ParsePeer(%q) = (%q, %q), want (%q, %q)", c.peer, peerId, peerName, c.wantId, c.wantName)
		}
	}
}
