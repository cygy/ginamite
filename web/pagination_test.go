package web

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestGetPageNumber(t *testing.T) {
	tests := []struct {
		query string
		want  int
	}{
		{"", 1},
		{"page=3", 3},
		{"page=abc", 1},
		{"page=0", 1},
		{"page=-2", 1},
		{"page=2&x=1", 2},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			c, _ := webtest.Context(http.MethodGet, "/?"+tt.query, nil)

			if got := GetPageNumber(c); got != tt.want {
				t.Errorf("GetPageNumber() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetPagination(t *testing.T) {
	tests := []struct {
		name         string
		items        int
		itemsPerPage int
		current      int
		wantCount    int
		wantPages    []int
	}{
		{"no item", 0, 10, 1, 0, []int{}},
		{"negative items", -5, 10, 1, 0, []int{}},
		{"no items per page", 100, 0, 1, 0, []int{}},
		{"negative items per page", 100, -10, 1, 0, []int{}},
		{"one partial page", 5, 10, 1, 1, []int{1}},
		{"exact pages", 20, 10, 1, 2, []int{1, 2}},
		{"partial last page", 21, 10, 1, 3, []int{1, 2, 3}},
		{"six pages are all displayed", 60, 10, 3, 6, []int{1, 2, 3, 4, 5, 6}},
		{"gap after the first pages", 100, 10, 1, 10, []int{1, 2, 3, -1, 8, 9, 10}},
		{"gap on both sides", 200, 10, 10, 20, []int{1, 2, 3, -1, 9, 10, 11, -1, 18, 19, 20}},
		{"gap before the last pages", 100, 10, 10, 10, []int{1, 2, 3, -1, 8, 9, 10}},
		{"current page next to the first ones", 200, 10, 4, 20, []int{1, 2, 3, 4, 5, -1, 18, 19, 20}},
		{"current page out of range", 100, 10, 50, 10, []int{1, 2, 3, -1, 8, 9, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPagination(tt.items, tt.itemsPerPage, tt.current)

			want := Pagination{
				CountOfPages: tt.wantCount,
				Pages:        tt.wantPages,
				TotalItems:   tt.items,
				ItemsPerPage: tt.itemsPerPage,
				CurrentPage:  tt.current,
				PreviousPage: tt.current - 1,
				NextPage:     tt.current + 1,
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("GetPagination() = %+v, want %+v", got, want)
			}
		})
	}
}
