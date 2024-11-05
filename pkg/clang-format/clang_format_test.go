package clang_format

import (
	"testing"
)

func Test_didLinesChange(t *testing.T) {
	type args struct {
		in map[string]int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "same list returns true",
			args: args{in: map[string]int{
				"foo": 5,
				"bar": 5,
				"baz": 5,
			}},
			want: true,
		},
		{
			name: "different returns false",
			args: args{in: map[string]int{
				"foo": 5,
				"bar": 5,
				"baz": 4,
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := didLinesChange(tt.args.in); got != tt.want {
				t.Errorf("didLinesChange() = %v, want %v", got, tt.want)
			}
		})
	}
}
