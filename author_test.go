// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

package main

import (
	"fmt"
	"testing"
)

func TestAuthor_Helper(t *testing.T) {
	testCases := []struct {
		author *Author
		want   string
	}{
		{&Author{"Épiste Olaire", "episte.olaire@mail.ma"}, "<address><a href=\"mailto:episte.olaire@mail.ma\">Épiste Olaire</a></address>"},
		{&Author{"", ""}, "<address><a href=\"mailto:\"></a></address>"},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.author.Helper(); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}
