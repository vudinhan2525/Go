package util

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPaginateFromRequest(ctx *gin.Context, defaultPage string, defaultLimit string) (int32, int32) {
	pageStr := ctx.DefaultQuery("page", defaultPage)
	limitStr := ctx.DefaultQuery("limit", defaultLimit)

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Invalid page value: %v", pageStr),
		})
		return 0, 0
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Invalid limit value: %v", limitStr),
		})
		return 0, 0
	}
	if (page < 1) || (limit < 1) {
		ctx.JSON(http.StatusInternalServerError, "Page and limit must be greater than 0")
		return 0, 0
	}
	return int32(page), int32(limit)
}
