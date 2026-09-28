package viper

import (
	"bytes"
	"reflect"
	"testing"
)

// Regression test for #2146: a nested mapping whose leaves are all YAML null
// must still be preserved by AllSettings (and therefore Unmarshal), rather
// than being dropped entirely.
func TestAllSettingsPreservesNilNested(t *testing.T) {
	data := "clients:\n  mysql:\n    protocol:\n    addr:\n"
	v := New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(data)); err != nil {
		t.Fatal(err)
	}

	all := v.AllSettings()
	clients, ok := all["clients"].(map[string]any)
	if !ok {
		t.Fatalf("clients missing from AllSettings: %#v", all)
	}
	mysql, ok := clients["mysql"].(map[string]any)
	if !ok {
		t.Fatalf("clients.mysql missing: %#v", clients)
	}
	want := map[string]any{"protocol": nil, "addr": nil}
	if !reflect.DeepEqual(mysql, want) {
		t.Fatalf("want %#v, got %#v", want, mysql)
	}
}
