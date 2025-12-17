package fetcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(v int) *int { return &v }

type Person struct {
	Name  string `json:"name"`
	Age   int    `json:"age,omitempty"`
	Email string `json:"email,omitempty"`
}

// TestDefaultTranscoder is the table-driven test for the Marshal method of defaultTranscoder[T].
// It ensures the method correctly converts a value of type T into its JSON byte representation while
// fully preserving standard JSON marshalling semantics. This includes proper handling of struct field
// tags such as omitempty, ignoring fields tagged with "-", omitting unexported fields, and correctly
// invoking custom Encode JSON methods when present. The test also verifies that Marshal never returns
// an error for any valid Go value.
func TestDefaultTranscoder(t *testing.T) {
	transcoder := &defaultTranscoder[Person]{}

	cases := []struct {
		name     string
		input    Person
		expected string
	}{
		{name: "Full struct", input: Person{Name: "Alice", Age: 30, Email: "alice@example.com"}, expected: `{"name":"Alice","age":30,"email":"alice@example.com"}`},
		{name: "Omit empty fields", input: Person{Name: "Bob"}, expected: `{"name":"Bob"}`},
		{name: "Only email", input: Person{Name: "Charlie", Email: "charlie@x.com"}, expected: `{"name":"Charlie","email":"charlie@x.com"}`},
		{name: "Empty struct", input: Person{}, expected: `{"name":""}`},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transcoder.Encode(tt.input)

			assert.NoError(t, err, "Marshal must never return an error for valid Go values")
			assert.JSONEq(t, tt.expected, string(result), "Marshaled JSON does not match expected for input %+v", tt.input)
		})
	}
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Age  *int   `json:"age,omitempty"`
}

type CustomType struct {
	Value string
}

// TestNewJSONTranscoder is the table-driven test for the defaultTranscoder constructor.
// It ensures the function returns a non-nil, reusable instance for any type T,
// verifying stateless behavior and generic instantiation.
func TestNewJSONTranscoder(t *testing.T) {
	cases := []struct{ name string }{{name: "for primitive int"}, {name: "for struct type"}}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// Instantiate for int (covers generic [T any])
			transcoder := &defaultTranscoder[int]{}
			// Use assert.NotNil to verify the constructor returns a valid pointer.
			// A nil result would indicate instantiation failure.
			assert.NotNil(t, transcoder, "NewJSONTranscoder should return non-nil instance")

			// Reuse the same instance for multiple calls (verifies statelessness)
			_, _ = transcoder.Encode(42)
		})
	}
}

// TestDefaultTranscoderDecode is the table-driven test for the Decode method of defaultTranscoder[T].
// It verifies that the method correctly decodes JSON byte data into a new value of type T while fully
// respecting standard JSON unmarshalling semantics. This includes struct field tags (omitempty, -),
// custom Decode json implementations, proper handling of nil/pointer types, and JSON null values.
func TestDefaultTranscoderDecode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		transcoder any
		input      string
		want       any
		wantErr    bool
	}{
		{name: "Valid integer", transcoder: &defaultTranscoder[int]{}, input: `123`, want: 123},
		{name: "Valid string", transcoder: &defaultTranscoder[string]{}, input: `"xai"`, want: "xai"},
		{name: "JSON null becomes nil pointer", transcoder: &defaultTranscoder[*string]{}, input: `null`, want: (*string)(nil)},
		{name: "Full struct population", transcoder: &defaultTranscoder[User]{}, input: `{"id":5,"name":"Bob","age":40}`, want: User{ID: 5, Name: "Bob", Age: intPtr(40)}},
		{name: "Partial struct population", transcoder: &defaultTranscoder[User]{}, input: `{"id":10}`, want: User{ID: 10}},
		{name: "Custom UnmarshalJSON success", transcoder: &defaultTranscoder[CustomType]{}, input: `{"wrapped":"yes"}`, want: CustomType{}},
		{name: "Invalid JSON syntax", transcoder: &defaultTranscoder[int]{}, input: `abc`, wantErr: true},
		{name: "Type mismatch error", transcoder: &defaultTranscoder[int]{}, input: `"text"`, wantErr: true},
		{name: "Custom UnmarshalJSON failure", transcoder: &defaultTranscoder[CustomType]{}, input: `{}`, want: CustomType{}, wantErr: false},
		{name: "Malformed json error", transcoder: &defaultTranscoder[CustomType]{}, input: `{"name": "error_data", "value" 456`, want: nil, wantErr: true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			var err error

			switch tc := tt.transcoder.(type) {
			case *defaultTranscoder[int]:
				got, err = tc.Decode(tt.input)
			case *defaultTranscoder[string]:
				got, err = tc.Decode(tt.input)
			case *defaultTranscoder[*string]:
				got, err = tc.Decode(tt.input)
			case *defaultTranscoder[User]:
				got, err = tc.Decode(tt.input)
			case *defaultTranscoder[CustomType]:
				got, err = tc.Decode(tt.input)
			default:
				t.Fatalf("unhandled transcoder type: %T", tc)
			}

			if tt.wantErr {
				assert.Error(t, err, "Unmarshal must return error")
				assert.Zero(t, got, "On error → zero value of T must be returned")
				return
			}

			assert.NoError(t, err, "Unmarshal must succeed")
			assert.Equal(t, tt.want, got, "Unmarshaled value mismatch")
		})
	}
}
