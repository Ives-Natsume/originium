package main

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

type operator struct {
	Name			string `json:"name"`
	Rarity			uint8  `json:"rarity"`	
}

var operators = []operator {
	{
		Name: "Typhon",
		Rarity: 6,
	},
	{
		Name: "Jessica",
		Rarity: 4,
	},
	{
		Name: "Angelina the Mellow Wish",
		Rarity: 6,
	},
	{
		Name: "Timeslot",
		Rarity: 5,
	},
}

func main() {
	router := gin.Default()
	router.GET("/operators", getOperators)
	router.GET("/operators/:rarity", getOperatorByRank)
	router.POST("/operators", postOperators)

	router.Run("localhost:8080")
}

func getOperators(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, operators)
}

func postOperators(c *gin.Context) {
	var newOp operator

	if err := c.ShouldBindJSON(&newOp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}


	operators = append(operators, newOp)
	c.IndentedJSON(http.StatusCreated, newOp)
}

func getOperatorByRank(c *gin.Context) {
	sRarity := c.Param("rarity")
	rarity64, err := strconv.ParseUint(sRarity, 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	rank := uint8(rarity64)
	result := make([]operator, 0)

	for _, op := range operators {
		if op.Rarity == rank {
			result= append(result, op)
		}
	}
	
	// if len(result) == 0 {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"message": "no operator found",
	// 	})
	// 	return
	// }

	c.IndentedJSON(http.StatusOK, result)
}