package policyutil

import (
	"testing"
)

type refCounter struct {
	count int
}

func (r refCounter) GetReferenceCount() int {
	return r.count
}

func TestGetReferenceCount(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  int
	}{
		{
			name:  "with GetReferenceCount() method",
			input: refCounter{count: 5},
			want:  5,
		},
		{
			name:  "without GetReferenceCount() method",
			input: "sample string",
			want:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := GetReferenceCount(test.input)
			if output != test.want {
				t.Errorf("want %d, got %d", test.want, output)
			}
		})
	}
}
