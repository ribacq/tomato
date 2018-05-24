package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCategory_mdTree(t *testing.T) {
	testCases := []struct {
		category  *Category
		prefix    string
		showPages bool
		want      string
	}{
		{&Category{Name: "Name"}, "p", false, "p* [Name >](/index.html)\n"},
		{&Category{Name: "Name", SubCategories: []*Category{{Name: "SubCat", Basename: "subcat", Parent: &Category{}}}}, "", false, "* [Name >](/index.html)\n\t* [SubCat >](/subcat/index.html)\n"},
		{&Category{Name: "Name", Pages: []*Page{{Basename: "index"}}}, "", true, "* [Name >](/index.html)\n"},
		{&Category{Name: "Name", Pages: []*Page{{Basename: "page", Title: "Page"}}}, "", true, "* [Name >](/index.html)\n\t* [Page](/page.html)\n"},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := string(tc.category.mdTree(tc.prefix, tc.showPages)); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestCategory_NavHelper(t *testing.T) {
	testCases := []struct {
		category  Category
		showPages bool
		want      string
	}{
		{Category{Name: "Name"}, false, "<ul>\n<li><a href=\"./index.html\">Name &gt;</a></li>\n</ul>\n"},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.category.NavHelper(Page{}, tc.showPages); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestCategory_FilterByTags(t *testing.T) {
	p0 := Page{Tags: []string{}}
	p1 := Page{Tags: []string{"a", "b"}}
	p2 := Page{Tags: []string{"a", "c"}}
	p3 := Page{Tags: []string{"d"}}

	testCases := []struct {
		category *Category
		tags     []string
		want     []*Page
	}{
		{&Category{}, nil, nil},
		{&Category{Pages: []*Page{&p0}}, nil, nil},
		{&Category{Pages: []*Page{&p0}}, []string{"a"}, nil},
		{&Category{Pages: []*Page{&p1, &p2, &p3}}, []string{"a"}, []*Page{&p1, &p2}},
		{&Category{Pages: []*Page{&p0, &p1}, SubCategories: []*Category{{Pages: []*Page{&p2, &p3}}}}, []string{"a"}, []*Page{&p1, &p2}},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.category.FilterByTags(tc.tags); reflect.DeepEqual(got, tc.want) == false {
				t.Errorf("got %v; want %v", got, tc.want)
			}
		})
	}
}

func TestCategory_PageCount(t *testing.T) {
	testCases := []struct {
		category *Category
		want     int
	}{
		{&Category{}, 0},
		{&Category{Pages: []*Page{{}, {}, {}, {}}}, 4},
		{&Category{Pages: []*Page{{}}, SubCategories: []*Category{{Pages: []*Page{{}}}}}, 2},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.category.PageCount(); got != tc.want {
				t.Errorf("got %d; want %d", got, tc.want)
			}
		})
	}
}

func TestCategory_CategoryCount(t *testing.T) {
	testCases := []struct {
		category *Category
		want     int
	}{
		{&Category{}, 0},
		{&Category{SubCategories: []*Category{{}, {}, {}, {}}}, 4},
		{&Category{SubCategories: []*Category{{SubCategories: []*Category{{}}}}}, 2},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.category.CategoryCount(); got != tc.want {
				t.Errorf("got %d; want %d", got, tc.want)
			}
		})
	}
}

func TestCategory_Path(t *testing.T) {
	testCases := []struct {
		category *Category
		want     string
	}{
		{&Category{}, "/"},
		{nil, "/"},
		{&Category{Parent: &Category{}, Basename: "test"}, "/test/"},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.category.Path(); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestCategory_Tags(t *testing.T) {
	testCases := []struct {
		category *Category
		want     []string
	}{
		{&Category{}, nil},
		{&Category{Pages: []*Page{{Tags: []string{"a", "b"}}}, SubCategories: []*Category{{Pages: []*Page{{Tags: []string{"b", "c"}}}}}}, []string{"a", "b", "c"}},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.category.Tags(); reflect.DeepEqual(got, tc.want) == false {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func TestCategory_RecentPages(t *testing.T) {
	p0 := Page{Date: "2042-04-02"}
	p1 := Page{Date: "2018-04-02"}
	p2 := Page{Date: "2017-04-02"}
	p3 := Page{Date: "2016-04-02"}
	p4 := Page{Date: "hello"}

	testCases := []struct {
		category *Category
		n        int
		want     []*Page
	}{
		{&Category{}, 0, nil},
		{&Category{}, 5, nil},
		{&Category{Pages: []*Page{&p3, &p2}, SubCategories: []*Category{{Pages: []*Page{&p1, &p0}}}}, 2, []*Page{&p0, &p1}},
		{&Category{Pages: []*Page{&p0, &p4}}, 2, []*Page{&p0, &p4}},
		{&Category{Pages: []*Page{&p4, &p0}}, 2, []*Page{&p0, &p4}},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.category.RecentPages(tc.n); reflect.DeepEqual(got, tc.want) == false {
				t.Errorf("got %v; want %v", got, tc.want)
			}
		})
	}
}

func TestCategory_FindParent(t *testing.T) {
	cat3 := Category{Basename: "cat3"}
	cat2 := Category{Basename: "cat2"}
	cat1 := Category{Basename: "cat1", SubCategories: []*Category{&cat2}}
	cat0 := Category{Basename: "cat0", SubCategories: []*Category{&cat1, &cat3}}

	testCases := []struct {
		fpath string
		want  *Category
		err   bool
	}{
		{":root:", nil, false},
		{"/", &cat0, false},
		{"/cat3/plop.html", &cat3, false},
		{"/cat1/cat2/hello.html", &cat2, false},
		{"/cat8/gnu.html", nil, true},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			got, err := cat0.FindParent(tc.fpath)
			if (err != nil) != tc.err {
				t.Errorf("got err = %v; want %v", err, tc.err)
			}
			if got != tc.want {
				t.Errorf("got category = %v; want %v", got, tc.want)
			}
		})
	}
}
