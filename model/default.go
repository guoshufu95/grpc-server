package model

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"strconv"
)

var ProductMap = map[string]Product{}

func Init() {
	for i := 1; i < 25; i++ {
		name := fmt.Sprintf("产品%s", strconv.Itoa(i))
		var status string
		if i%2 == 0 {
			status = "1"
		} else {
			status = "2"
		}
		ProductMap[strconv.Itoa(i)] = Product{
			Id:     i,
			Name:   name,
			Status: status,
		}
	}
}

type MyClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type Product struct {
	Id     int    `json:"productId"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
