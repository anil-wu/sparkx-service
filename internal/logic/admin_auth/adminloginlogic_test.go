package admin_auth

import "testing"

func TestPasswordMatches(t *testing.T) {
	t.Run("stored md5 lower input plain", func(t *testing.T) {
		if !passwordMatches("21232f297a57a5a743894a0e4a801fc3", "admin") {
			t.Fatal("expected match")
		}
	})

	t.Run("stored md5 upper input plain", func(t *testing.T) {
		if !passwordMatches("21232F297A57A5A743894A0E4A801FC3", "admin") {
			t.Fatal("expected match")
		}
	})

	t.Run("stored md5 lower input md5 upper", func(t *testing.T) {
		if !passwordMatches("21232f297a57a5a743894a0e4a801fc3", "21232F297A57A5A743894A0E4A801FC3") {
			t.Fatal("expected match")
		}
	})

	t.Run("stored plain input plain", func(t *testing.T) {
		if !passwordMatches("admin", "admin") {
			t.Fatal("expected match")
		}
	})

	t.Run("stored plain input mismatch", func(t *testing.T) {
		if passwordMatches("admin", "admin1") {
			t.Fatal("expected mismatch")
		}
	})
}

func TestStatusAllowsLogin(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{"", true},
		{"active", true},
		{"ACTIVE", true},
		{"enabled", true},
		{"1", true},
		{"true", true},
		{"disabled", false},
		{"inactive", false},
		{"0", false},
		{"false", false},
		{"other", false},
	}

	for _, tc := range cases {
		if got := statusAllowsLogin(tc.status); got != tc.want {
			t.Fatalf("statusAllowsLogin(%q)=%v, want %v", tc.status, got, tc.want)
		}
	}
}

