package web

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPageNumber : returns the page number from the request.
func GetPageNumber(c *gin.Context) int {
	var pageNumber int
	var err error
	if pageNumber, err = strconv.Atoi(c.Query("page")); err != nil {
		pageNumber = 1
	}

	return pageNumber
}

// GetPagination : returns an initialized Pagination struct.
func GetPagination(countOfItems, itemsPerPage, currentPage int) Pagination {
	countOfPages := countOfItems / itemsPerPage
	if countOfItems%itemsPerPage > 0 {
		countOfPages++
	}

	pages := []int{}
	for i := 1; i <= countOfPages; i++ {
		if i <= 3 || i >= countOfPages-2 || (i >= (currentPage-1) && i <= (currentPage+1)) {
			pages = append(pages, i)
		} else {
			if pages[len(pages)-1] != -1 {
				pages = append(pages, -1)
			}
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
