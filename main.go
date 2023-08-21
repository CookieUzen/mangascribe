package main

import (
	"flag"
	"github.com/CookieUzen/mangascribe/DB"
	"github.com/CookieUzen/mangascribe/Config"
	"github.com/CookieUzen/mangascribe/Models"
	"github.com/golang/glog"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/CookieUzen/mangascribe/docs"
	"net/http"
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

	v1 := r.Group("/v1") 
	v1.GET("/meow", func(c *gin.Context) {
		c.JSON(418, gin.H {
			"message": "Teapot",
		})
	})
	v1.POST("/accounts", func(c *gin.Context) {registerHandler(c, &dbm)})

	v1.POST("/login", func(c *gin.Context) {loginHandler(c, &dbm)})
	// TODO: figure out how to update account info
	// TODO: revoke API keys and an API key endpoint

	// This endpoint serves the Swagger UI and the OpenAPI spec
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Run the server
	r.Run(Config.GIN_URL+":"+Config.GIN_PORT)

	// Flush logs
	glog.Flush()

	// Close the database connection
	dbm.Close()
}

// registerHandler Register a new account
// @Summary Register a new account
// @Description register a new account by json user
// @Tags user
// @Accept  json
// @Produce  json
// @Param accountInfo body Models.NewAccountRequest true "Account information for registration"
// @Success 200 {object} Models.Response_APIKey
// @Failure 400,502 {object} Models.Fail
// @Router /v1/accounts [post]
func registerHandler(c *gin.Context, dbm *DB.DBManager) {
	var form Models.NewAccountRequest

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, Models.Fail{Error: err.Error()})
		return
	}

	account, err := dbm.CreateAccount(form)
	if err != nil {
		c.JSON(http.StatusBadRequest, Models.Fail{Error: err.Error()})
		return
	}

	api_key, err := dbm.GenerateAPIKey(account, Config.DEFAULT_API_KEY_EXPIRATION)
	if err != nil {
		c.JSON(http.StatusBadGateway, Models.Fail{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Models.Response_APIKey{APIKey: api_key.ToJSON()})
}

// loginHandler Login to a user account and return all API keys associated with the account
// @Summary Login a user
// @Description login user by json user
// @Tags user
// @Accept  json
// @Produce  json
// @Param user body Models.LoginRequest true "Login user credentials"
// @Success 200 {object} Models.Response_APIKeyList
// @Failure 502,400,401 {object} Models.Fail
// @Router /v1/login [post]
func loginHandler(c *gin.Context, dbm *DB.DBManager) {
	var form Models.LoginRequest

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, Models.Fail{Error: err.Error()})
		return
	}

	var account Models.Account
	success, err := dbm.AuthAccount(&account, form)
	if err != nil {
		c.JSON(http.StatusBadGateway, Models.Fail{Error: err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusUnauthorized, Models.Fail{Error: "Invalid username or password"})
		return
	}

	api_keys, err := dbm.GetAPIKeys(&account)
	if err != nil {
		c.JSON(http.StatusBadGateway, Models.Fail{Error: err.Error()})
		return
	}

	json_keys := make([]Models.APIKeyJSON, len(api_keys))

	// Loop through the API keys converting them to JSON
	for i, key := range api_keys {
		json_keys[i] = key.ToJSON()
	}


	c.JSON(http.StatusOK, Models.Response_APIKeyList{APIKeys: json_keys})
}


// generateAPIKey Generate a new API key for an account
// @Summary Generate a new API key for an account
// @Description generate a new API key for an account
// @Tags user
// @Accept  json
// @Produce  json
// @Param accountInfo body Models.LoginRequest true "Login user credentials"
// @Success 200 {object} Models.Response_APIKey
// @Failure 502,400,401 {object} Models.Fail
// @Router /v1/accounts [get]
func generateAPIKey(c *gin.Context, dbm *DB.DBManager) {
	var form Models.LoginRequest

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, Models.Fail{Error: err.Error()})
		return
	}

	var account Models.Account
	success, err := dbm.AuthAccount(&account, form)
	if err != nil {
		c.JSON(http.StatusBadGateway, Models.Fail{Error: err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusUnauthorized, Models.Fail{Error: "Invalid username or password"})
		return
	}

	api_key, err := dbm.GenerateAPIKey(&account, Config.DEFAULT_API_KEY_EXPIRATION)
	if err != nil {
		c.JSON(http.StatusBadGateway, Models.Fail{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Models.Response_APIKey{APIKey: api_key.ToJSON()})
}
