package terminal

import (
	"strings"
	"testing"
)

func TestAll(t *testing.T) {
	for _, test := range []struct {
		set         string
		wantErr     string
		stoppers    []int
		nonStoppers []int
	}{
		{
			set:         "100",
			wantErr:     "",
			stoppers:    []int{100, 101, 199},
			nonStoppers: []int{200, 300, 400, 500},
		},
		{
			set:         "100,200",
			wantErr:     "",
			stoppers:    []int{100, 101, 199, 200, 201, 299},
			nonStoppers: []int{300, 400, 500},
		},
		{
			set:         "100,300",
			wantErr:     "",
			stoppers:    []int{100, 101, 199, 300, 301, 399},
			nonStoppers: []int{200, 400, 500},
		},
		{
			set:         "300,100",
			wantErr:     "",
			stoppers:    []int{100, 101, 199, 300, 301, 399},
			nonStoppers: []int{200, 400, 500},
		},
		{
			set:         "300,100,sausage",
			wantErr:     "not a number",
			stoppers:    []int{100, 101, 199, 300, 301, 399},
			nonStoppers: []int{200, 400, 500},
		},
		{
			set:         "300,101",
			wantErr:     "not a round hundred",
			stoppers:    []int{100, 101, 199, 300, 301, 399},
			nonStoppers: []int{200, 400, 500},
		},
	} {
		term, err := New(test.set)
		switch {
		case err == nil && test.wantErr != "":
			t.Errorf("New(%q) = _,nil, want error with %q", test.set, test.wantErr)
		case err != nil && test.wantErr == "":
			t.Errorf("New(%q) = _,%q, want nil error", test.set, err.Error())
		case err != nil && test.wantErr != "" && !strings.Contains(err.Error(), test.wantErr):
			t.Errorf("New(%q) = _,%q, want error with %q", test.set, err.Error(), test.wantErr)
		case err == nil && test.wantErr == "":
			for _, v := range test.stoppers {
				if term.IsStop(v) != true {
					t.Errorf("%q: IsStop(%v) = false, want true", test.set, v)
				}
			}
			for _, v := range test.nonStoppers {
				if term.IsStop(v) != false {
					t.Errorf("%q: IsStop(%v) = true, want false", test.set, v)
				}
			}
		}
	}
}
