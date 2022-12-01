package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"

	gecko "github.com/superoo7/go-gecko/v3"
)

var redisClient *redis.Client
var redisTimeout int

type Coin struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	connectRedis()
	redis_Timeout := os.Getenv("REDIS_TIMEOUT")
	var err error
	redisTimeout, err = strconv.Atoi(redis_Timeout)
	if err != nil {
		redisTimeout = 5
	}
	r := gin.Default()
	r.GET("/price", GetPrice)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	err = r.Run(":" + port)
	if err != nil {
		log.Fatalf("impossible to start server: %s", err)
	}
}

func GetPrice(c *gin.Context) {
	name := c.Query("name")
	var err error
	var coin Coin
	redisValue, err := redisClient.Get(name).Result()
	if err == nil && redisValue != "" {
		redisIntVal, _ := strconv.Atoi(redisValue)
		coin = Coin{
			Name:  name,
			Price: float64(redisIntVal),
		}
	} else {
		coin, err = GetCoinPrice(name)
		setToRedis(name, strconv.Itoa(int(coin.Price)))
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible to retrieve bitcoin price" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, coin)
}

func GetCoinPrice(name string) (Coin, error) {
	cg := gecko.NewClient(nil)
	coin, err := cg.CoinsID(name, true, true, true, true, true, true)
	if err != nil {
		return Coin{}, err
	}
	return Coin{
		Name:  name,
		Price: coin.MarketData.CurrentPrice["usd"],
	}, nil
}

func connectRedis() {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pong)

	redisClient = client
}

func setToRedis(key, val string) {
	err := redisClient.Set(key, val, time.Duration(redisTimeout*60)).Err()
	if err != nil {
		fmt.Println(err)
	}
}
