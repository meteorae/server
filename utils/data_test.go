package utils_test

import (
	"testing"

	"github.com/meteorae/meteorae-server/utils"
)

func TestIsStringInSlice(t *testing.T) {
	t.Parallel()

	type args struct {
		a    string
		list []string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "String is in slice",
			args: args{
				a:    "a",
				list: []string{"a", "b", "c"},
			},
			want: true,
		},
		{
			name: "String is not in slice",
			args: args{
				a:    "d",
				list: []string{"a", "b", "c"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := utils.IsStringInSlice(tt.args.a, tt.args.list); got != tt.want {
				t.Errorf("IsStringInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
