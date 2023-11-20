package endpoints

import (
	"testing"
)

func TestNew(t *testing.T) {
	// This doesn't test the underlying url.Parse() actions and if they return errors.
	for _, test := range []struct {
		set     string
		wantErr bool
	}{
		{
			set:     "http://localhost:1234,http://localhost:5678",
			wantErr: false,
		},
		{
			set:     "https://localhost:1234,https://localhost:5678",
			wantErr: false,
		},
		{
			set:     "https://localhost:1234/path,https://localhost:5678/another/path",
			wantErr: false,
		},
		{
			set:     "https://localhost:1234/path,",
			wantErr: true,
		},
	} {
		_, err := New(test.set)
		gotErr := err != nil
		if gotErr != test.wantErr {
			t.Errorf("New(%q): got error = %v, want error = %v", test.set, gotErr, test.wantErr)
		}
	}
}
