package toJson

import (
	"testing"
	"reflect"
	"errors"
)

var addresser = "ss"

type s struct {
	A string
	B string
	c bool
}

var jsontests = []struct {
	in  interface{}
	out interface{}
}{
	{ "s", "s" },
	{ 1.0, 1.0 },
	{ 1, 1 },
	{nil, nil},
	{(*string)(nil), nil},
	{&addresser, "ss"},
	{[]interface{}{"a", "b", "c"}, []interface{}{"a", "b", "c"}},
	{[]interface{}{"a", 2, true}, []interface{}{"a", 2, true}},
	{errors.New("foo"), "foo"},
	{s{"foo", "foo", true}, map[string]interface{}{"a":"foo", "b":"foo"}},
	{[]interface{}{s{"foo", "foo", false}, s{"foo", "bar", false}}, []interface{}{map[string]interface{}{"a":"foo", "b":"foo"}, map[string]interface{}{"a":"foo", "b":"bar"}}},
}

func TestToJson(t *testing.T) {
	for _, tt := range jsontests {
		output, _ := ToJson(tt.in)

		var is_error bool

		if !reflect.DeepEqual(output, tt.out) {
			is_error = true
		}

		if is_error {
			t.Errorf("ToJson(%v) => %v, want %v", tt.in, output, tt.out)
		}
	}
}
