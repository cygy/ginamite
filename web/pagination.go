package web

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Pagination holds the properties of a paginated list.
type Pagination struct {
	CountOfPages int
	Pages        []int // The numbers of the pages to display, where -1 stands for the pages left out.
	TotalItems   int
	ItemsPerPage int
	CurrentPage  int
	PreviousPage int
	NextPage     int
}

// GetPageNumber returns the number of the page in the query parameter "page" of the request: 1 if it is missing,
// invalid, or lower than 1.
func GetPageNumber(c *gin.Context) int {
	pageNumber, err := strconv.Atoi(c.Query("page"))
	if err != nil || pageNumber < 1 {
		return 1
	}

	return pageNumber
}

// GetPagination returns the pagination of countOfItems items, itemsPerPage per page, on the page currentPage.
// The pages displayed are the first three, the last three and the ones next to the current one, separated by -1.
// There is no page if there is no item or if itemsPerPage is lower than 1.
func GetPagination(countOfItems, itemsPerPage, currentPage int) Pagination {
	countOfPages := 0
	if countOfItems > 0 && itemsPerPage > 0 {
		countOfPages = (countOfItems + itemsPerPage - 1) / itemsPerPage
	}

	pages := []int{}
	for i := 1; i <= countOfPages; i++ {
		switch {
		case i <= 3 || i >= countOfPages-2 || (i >= currentPage-1 && i <= currentPage+1):
			pages = append(pages, i)
		case pages[len(pages)-1] != -1:
			pages = append(pages, -1)
		}
	}

	return Pagination{
		CountOfPages: countOfPages,
		Pages:        pages,
		TotalItems:   countOfItems,
		ItemsPerPage: itemsPerPage,
		CurrentPage:  currentPage,
		PreviousPage: currentPage - 1,
		NextPage:     currentPage + 1,
	}
}
