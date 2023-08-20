package main

import (
	"flag"
	"fmt"
	"github.com/CookieUzen/mangascribe/DB"
	// "github.com/CookieUzen/mangascribe/Config"
	"github.com/CookieUzen/mangascribe/Models"
	"github.com/golang/glog"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/CookieUzen/mangascribe/docs"
	"time"
)

// TODO finetune -v levels
// TODO fix glog error passing

//	@title			Mangascribe API
//	@description	This is a mangascribe API server.
//	@version		1.0
//	@host			localhost:8080
func main() {
	// For logging flags
	flag.Parse()

	// Connect to the database
	dbm := DB.Open()

	// Set up gin server
	r := gin.Default()

	// This endpoint serves the Swagger UI and the OpenAPI spec
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/v1") 
	v1.GET("/meow", func(c *gin.Context) {
		c.JSON(418, gin.H {
			"message": "Teapot",
		})
	})

	// r.Run(Config.GIN_URL+":"+Config.GIN_PORT)

	// Create a new account
	userRequest := Models.NewAccountRequest{
		Username: "CookieUzen",
		Password: "password",
		Email: "uzen@cookieuz.io",
	}

	user, err := dbm.CreateAccount(userRequest)
	if err != nil {
		glog.Error(err)
	}
	fmt.Println(user)

	// Create a new API key
	key, err := dbm.GenerateAPIKey(user, time.Hour * 24 * 7)
	if err != nil {
		glog.Error(err)
	}

	expKey, err := dbm.GenerateAPIKey(user, time.Duration(0))
	if err != nil {
		glog.Error(err)
	}

	// Get the user for testing
	var user2 Models.Account
	_, err = dbm.AuthAccount(&user2, Models.LoginRequest{
		Identifier: "CookieUzen",
		Password: "password",
	})
	fmt.Println(user2)

	// Get the user from the API key
	var user3 Models.Account
	err = dbm.UserFromKey(&user3, key)
	fmt.Println(user3)

	// Get the user from the expired API key
	var user4 Models.Account
	err = dbm.UserFromKey(&user4, expKey)
	glog.Error(err)
	fmt.Println(user4)

	// Flush logs
	glog.Flush()

	// Close the database connection
	dbm.Close()
}
