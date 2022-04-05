package utils

import (
	"reflect"
	"testing"
)

func TestGetResultsFromRegex(t *testing.T) {
	type args struct {
		pattern string
		tmpl    string
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "correct_result",
			args: args{
				pattern: `^my-app-([0-9]+\.[0-9]+\.[0-9]+)$`,
				tmpl:    `$1`,
				content: `my-app-0.0.1`,
			},
			want:    "0.0.1",
			wantErr: false,
		},
		{
			name: "empty_result",
			args: args{
				pattern: `^my-app-([0-9]+\.[0-9]+\.[0-9]+)$`,
				tmpl:    `$1`,
				content: `my-app-0.0.1-SNAPSHOT`,
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResultsFromRegex(tt.args.pattern, tt.args.tmpl, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResultsFromRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetResultsFromRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	type args struct {
		pattern string
		tmpl    string
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   string
		wantErr bool
	}{
		{
			name: "should_match",
			args: args{
				pattern: `^my-app-([0-9]+\.[0-9]+\.[0-9]+)$`,
				tmpl:    `$1`,
				version: "my-app-0.0.1",
			},
			want:    true,
			want1:   "0.0.1",
			wantErr: false,
		},
		{
			name: "should_not_match",
			args: args{
				pattern: `^my-app-([0-9]+\.[0-9]+\.[0-9]+)$`,
				tmpl:    `$1`,
				version: "my-app-0.0.1-SNAPSHOT",
			},
			want:    false,
			want1:   "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := MatchPattern(tt.args.pattern, tt.args.tmpl, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchPattern() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("MatchPattern() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMeetConstraint(t *testing.T) {
	type args struct {
		constraint string
		ver        string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "meet_constraint",
			args: args{
				constraint: ">=0.0.1",
				ver:        "0.0.1",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not_meet_constraint",
			args: args{
				constraint: ">=0.0.1",
				ver:        "0.0.0",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MeetConstraint(tt.args.constraint, tt.args.ver)
			if (err != nil) != tt.wantErr {
				t.Errorf("MeetConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MeetConstraint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	type args struct {
		l []string
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "contains",
			args: args{
				l: []string{"a", "b", "c"},
				s: "a",
			},
			want: true,
		},
		{
			name: "not_contains",
			args: args{
				l: []string{"a", "b", "c"},
				s: "d",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.l, tt.args.s); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsInt(t *testing.T) {
	type args struct {
		l []int
		i int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "contains",
			args: args{
				l: []int{1, 2, 3},
				i: 1,
			},
			want: true,
		},
		{
			name: "not_contains",
			args: args{
				l: []int{1, 2, 3},
				i: 4,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsInt(tt.args.l, tt.args.i); got != tt.want {
				t.Errorf("ContainsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveDuplicateStr(t *testing.T) {
	type args struct {
		strSlice []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "remove_duplicate",
			args: args{
				strSlice: []string{"a", "b", "c", "a", "d", "c"},
			},
			want: []string{"a", "b", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveDuplicateStr(tt.args.strSlice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicateStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
