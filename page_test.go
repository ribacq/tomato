package main

import (
	"fmt"
	"testing"
)

func TestPage_ContentHelper(t *testing.T) {
	testCases := []struct {
		page Page
		want string
	}{
		{Page{Content: []byte("page [test](test)")}, "<p>page <a href=\"test\">test</a></p>\n"},
		{Page{}, ""},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.page.ContentHelper(); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestPage_PathHelper(t *testing.T) {
	testCases := []struct {
		page Page
		want string
	}{
		{Page{Basename: "index"}, ""},
		{Page{Basename: "index", Category: &Category{}}, "<a href=\"./index.html\"></a>"},
		{Page{Basename: "index", Category: &Category{Name: "Category"}}, "<a href=\"./index.html\">Category</a>"},
		{Page{Basename: "test", Title: "Test"}, "<a href=\"./test.html\">Test</a>"},
		{Page{Basename: "test", Title: "Test", Category: &Category{}}, "<a href=\"./index.html\"></a> &gt; <a href=\"./test.html\">Test</a>"},
		{Page{Basename: "test", Title: "Test", Category: &Category{Name: "Category"}}, "<a href=\"./index.html\">Category</a> &gt; <a href=\"./test.html\">Test</a>"},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.page.PathHelper(tc.page); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestPage_Path(t *testing.T) {
	testCases := []struct {
		page *Page
		want string
	}{
		{&Page{}, "/.html"},
		{&Page{Basename: "page"}, "/page.html"},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.page.Path(); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestPage_PathToRoot(t *testing.T) {
	testCases := []struct {
		page *Page
		want string
	}{
		{&Page{}, "."},
		{&Page{Category: &Category{}}, "."},
		{&Page{Category: &Category{Parent: &Category{}}}, "./.."},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.page.PathToRoot(); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}
