package str

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubString(t *testing.T) {
	type args struct {
		str   string
		start int
		end   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test substring start and end",
			args: args{str: "The lazy brown fox jumps the fence", start: 2, end: 5},
			want: "e l",
		}, {
			name: "Test substring start",
			args: args{str: "The lazy brown fox jumps the fence", start: 2, end: 500},
			want: "e lazy brown fox jumps the fence",
		}, {
			name: "Test substring end",
			args: args{str: "The lazy brown fox jumps the fence", start: 0, end: 5},
			want: "The l",
		}, {
			name: "Test substring returns empty for negative start",
			args: args{str: "The lazy brown fox jumps the fence", start: -1, end: 5},
			want: "",
		}, {
			name: "Test substring returns empty for negative end",
			args: args{str: "The lazy brown fox jumps the fence", start: 0, end: -1},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(
					t,
					tt.want,
					SubString(tt.args.str, tt.args.start, tt.args.end),
					"SubString(%v, %v, %v)",
					tt.args.str,
					tt.args.start,
					tt.args.end)
			})
	}
}
