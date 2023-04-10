package main

import (
	"fmt"
	"net/http"

	"github.com/asdv23/go-binance/binanceapi"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	apiKey    = "your api key"
	secretKey = "your secret key"
	port      = ":3000"
)

func main() {
	loadConfig()
	r := gin.Default()

	fapi := binanceapi.NewFuturesAPI(apiKey, secretKey)
	r.GET("/fetch-trades", fapi.FetchTrades)
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/btc", fapi.BTC)
	r.GET("/allfut", fapi.Allfut)

	r.Run(port) // listen and serve on 0.0.0.0:3000 (for windows "localhost:3000")
}

func loadConfig() {
	viper.SetConfigName("config")   // name of config file (without extension)
	viper.SetConfigType("yaml")     // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("./config") // optionally look for config in the working directory
	err := viper.ReadInConfig()     // Find and read the config file
	if err != nil {                 // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	apiKey = viper.GetString("server.apiKey")
	secretKey = viper.GetString("server.secretKey")
	port = viper.GetString("server.port")
}
