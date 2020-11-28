package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type A struct {
	a string `json:"a" binding:"a"`
}

func main() {
	r := gin.Default()
	r.POST("/a", test)
	r.Run(":8456")
}

func test(ctx *gin.Context) {
	fmt.Println(ctx)
	var a A
	err := ctx.ShouldBindJSON(&a)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(a)
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "success",
	})
}
