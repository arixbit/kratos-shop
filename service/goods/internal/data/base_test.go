package data

import (
	"reflect"
	"testing"
)

func TestGormListScan(t *testing.T) {
	var list GormList
	if err := list.Scan([]byte(`["a","b"]`)); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]string(list), []string{"a", "b"}) {
		t.Fatalf("unexpected list: %v", list)
	}

	if err := list.Scan(`["c"]`); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]string(list), []string{"c"}) {
		t.Fatalf("unexpected list after string scan: %v", list)
	}

	if err := list.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("nil scan should produce empty list: %v", list)
	}
}

func TestGormListValue(t *testing.T) {
	list := GormList{"x", "y"}
	v, err := list.Value()
	if err != nil {
		t.Fatal(err)
	}
	if string(v.([]byte)) != `["x","y"]` {
		t.Fatalf("unexpected value: %s", v)
	}
}
