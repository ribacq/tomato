package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSiteinfo_MainAuthorHelper(t *testing.T) {
	testCases := []struct {
		siteinfo Siteinfo
		want     string
	}{
		{Siteinfo{Authors: []Author{{"A", "a"}, {"B", "b"}}}, "<address><a href=\"mailto:a\">A</a></address>"},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.siteinfo.MainAuthorHelper(); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestSiteinfo_CopyrightHelper(t *testing.T) {
	testCases := []struct {
		siteinfo Siteinfo
		want     string
	}{
		{Siteinfo{Copyright: "test [test](test)"}, "<p>test <a href=\"test\">test</a></p>\n"},
		{Siteinfo{}, ""},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.siteinfo.CopyrightHelper(&Page{}); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestSiteinfo_SubtitleHelper(t *testing.T) {
	testCases := []struct {
		siteinfo Siteinfo
		want     string
	}{
		{Siteinfo{Subtitle: "sous-titre [test](test)"}, "<p>sous-titre <a href=\"test\">test</a></p>\n"},
		{Siteinfo{}, ""},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.siteinfo.SubtitleHelper(&Page{}); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestSiteinfo_DescriptionHelper(t *testing.T) {
	testCases := []struct {
		siteinfo Siteinfo
		want     string
	}{
		{Siteinfo{Description: "description [test](test)"}, "<p>description <a href=\"test\">test</a></p>\n"},
		{Siteinfo{}, ""},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.siteinfo.DescriptionHelper(&Page{}); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestSiteinfo_FindAuthor(t *testing.T) {
	author := Author{Name: "episte"}

	testCases := []struct {
		siteinfo   *Siteinfo
		want       *Author
		name       string
		shouldFail bool
	}{
		{&Siteinfo{}, nil, "", true},
		{&Siteinfo{Authors: []Author{author}}, &author, "episte", false},
	}
	for ti, tc := range testCases {
		t.Run(fmt.Sprintf("%d", ti), func(t *testing.T) {
			if got, err := tc.siteinfo.FindAuthor(tc.name); !tc.shouldFail && err != nil || tc.shouldFail && err == nil || reflect.DeepEqual(got, tc.want) == false {
				if tc.shouldFail {
					t.Errorf("got %s, %s; want %s, a non-nil error", got, err, tc.want)
				} else {
					t.Errorf("got %s, %s; want %s, nil", got, err, tc.want)
				}
			}
		})
	}
}
