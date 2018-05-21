package main

import (
	"fmt"
	"reflect"
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
			if got := tc.siteinfo.CopyrightHelper(Page{}); got != tc.want {
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
			if got := tc.siteinfo.SubtitleHelper(Page{}); got != tc.want {
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
			if got := tc.siteinfo.DescriptionHelper(Page{}); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

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

	testCases := []struct {
		category *Category
		n        int
		want     []*Page
	}{
		{&Category{}, 0, nil},
		{&Category{}, 5, nil},
		{&Category{Pages: []*Page{&p3, &p2}, SubCategories: []*Category{{Pages: []*Page{&p1, &p0}}}}, 2, []*Page{&p0, &p1}},
	}
	for tci, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tci), func(t *testing.T) {
			if got := tc.category.RecentPages(tc.n); reflect.DeepEqual(got, tc.want) == false {
				t.Errorf("got %v; want %v", got, tc.want)
			}
		})
	}
}
