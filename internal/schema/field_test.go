package schema

import "testing"

func TestFieldFindChild(t *testing.T) {
	t.Parallel()

	root := &Field{
		Name: "root",
		Children: []*Field{
			{Name: "server", Type: FieldTypeStruct},
			nil,
			{Name: "app", Type: FieldTypeStruct},
		},
	}

	got, ok := root.FindChild("app")
	if !ok {
		t.Fatal("FindChild() did not find existing child")
	}
	if got.Name != "app" {
		t.Fatalf("FindChild() returned %q, want %q", got.Name, "app")
	}
}

func TestFieldFindChildMissing(t *testing.T) {
	t.Parallel()

	root := &Field{
		Name:     "root",
		Children: []*Field{{Name: "server", Type: FieldTypeStruct}},
	}

	got, ok := root.FindChild("missing")
	if ok {
		t.Fatal("FindChild() found missing child")
	}
	if got != nil {
		t.Fatalf("FindChild() returned %#v, want nil", got)
	}
}

func TestFieldFindChildNilReceiver(t *testing.T) {
	t.Parallel()

	var root *Field
	got, ok := root.FindChild("app")
	if ok {
		t.Fatal("FindChild() found child on nil receiver")
	}
	if got != nil {
		t.Fatalf("FindChild() returned %#v, want nil", got)
	}
}

func TestFieldTypeIsScalar(t *testing.T) {
	t.Parallel()

	tests := map[FieldType]bool{
		FieldTypeString:  true,
		FieldTypeBool:    true,
		FieldTypeInt:     true,
		FieldTypeFloat:   true,
		FieldTypeStruct:  false,
		FieldTypeSlice:   false,
		FieldTypeArray:   false,
		FieldTypeMap:     false,
		FieldTypeUnknown: false,
	}

	for fieldType, want := range tests {
		fieldType := fieldType
		want := want
		t.Run(string(fieldType), func(t *testing.T) {
			t.Parallel()

			if got := fieldType.IsScalar(); got != want {
				t.Fatalf("IsScalar() = %t, want %t", got, want)
			}
		})
	}
}
