package domain

import (
	"reflect"
	"testing"
)

func TestNormalizeDigestSectionOrder(t *testing.T) {
	got := NormalizeDigestSectionOrder([]string{"emails", "bogus", "highlights", "emails"})
	want := []string{"emails", "highlights", "topics"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	if !reflect.DeepEqual(NormalizeDigestSectionOrder(nil), DefaultDigestSectionOrder) {
		t.Fatalf("nil order should become default")
	}
}

func TestNormalizeDigestDensity(t *testing.T) {
	if got := NormalizeDigestDensity("compact"); got != DigestDensityCompact {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeDigestDensity("weird"); got != DigestDensityComfortable {
		t.Fatalf("got %q", got)
	}
}
