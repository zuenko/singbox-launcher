package subscription

import (
	"reflect"
	"testing"
)

func TestHysteria2MportSpecToSingBoxServerPorts(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"41000", []string{"41000:41000"}},
		{"41000-42000", []string{"41000:42000"}},
		{"41000:42000", []string{"41000:42000"}},
		{"443,20000-30000", []string{"443:443", "20000:30000"}},
		{" 41000 , 42000 ", []string{"41000:41000", "42000:42000"}},
		{"", nil},
		{"   ,  ", nil},
	}
	for _, tt := range tests {
		got := hysteria2MportSpecToSingBoxServerPorts(tt.in)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("hysteria2MportSpecToSingBoxServerPorts(%q) = %#v, want %#v", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeHysteria2ServerPortsSlice(t *testing.T) {
	got := NormalizeHysteria2ServerPortsSlice([]string{"41000", "20000-30000"})
	want := []string{"41000:41000", "20000:30000"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestHysteria2RecoverMultiPortAuthority(t *testing.T) {
	raw := "hysteria2://secret@example.com:443,20000-30000/?insecure=1#n"
	u, plist, err := hysteria2RecoverMultiPortAuthority(raw)
	if err != nil {
		t.Fatal(err)
	}
	if plist != "443,20000-30000" {
		t.Fatalf("plist %q", plist)
	}
	if u.Hostname() != "example.com" || u.Port() != "443" {
		t.Fatalf("host/port: %+v", u)
	}
}
