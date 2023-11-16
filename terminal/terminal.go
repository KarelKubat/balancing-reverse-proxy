// Package terminal determines which HTTP statuses we should consider 'terminal' in the sense that
// an endpoint's response can be taken and returned to the caller.
package terminal

import (
	"fmt"
	"strconv"
	"strings"
)

type Terminal struct {
	hundreds map[int]struct{}
}

func New(set string) (*Terminal, error) {
	t := &Terminal{
		hundreds: map[int]struct{}{},
	}
	for _, part := range strings.Split(set, ",") {
		start, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("terminal statuses: part %q of %q is not a number", part, set)
		}
		if start%100 != 0 {
			return nil, fmt.Errorf("terminal statuses: part %q of %q is not a round hundred", part, set)
		}
		t.hundreds[start] = struct{}{}
	}
	return t, nil
}

func (t *Terminal) IsStop(st int) bool {
	st -= st % 100
	_, ok := t.hundreds[st]
	return ok
}
