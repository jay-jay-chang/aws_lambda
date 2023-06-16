package main

import (
	// go內建
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	// package別名
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	// gin框架
	"github.com/gin-gonic/gin"
	//redis
	redis "github.com/redis/go-redis/v9"
)

var ginLambda *ginadapter.GinLambda

func init() {
	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Gin code start")
	// instance gin
	r := gin.Default()
	//設定一個route方法
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	//設定第二個
	r.GET("/api1", api1)

	//測試redis client
	r.GET("/test_redis", func(c *gin.Context) {
		TestRedisClient()
		c.JSON(200, gin.H{
			"message": "test redis",
		})
	})

	// instance GinLambda object
	ginLambda = ginadapter.New(r)
}

var redis_ctx = context.Background()

// redis
func TestRedisClient() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis-test.y1emey.ng.0001.apne1.cache.amazonaws.com:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := rdb.Set(redis_ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get(redis_ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := rdb.Get(redis_ctx, "key2").Result()
	//redis.Nil為string, 找不到key時會回傳
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist
}

func api1(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong1",
	})
}

// Handler 因lambda只支援特定參數格式，請參照https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html
// context.Context - https://docs.aws.amazon.com/lambda/latest/dg/golang-context.html
// request以及response的josn object要符合lamda規則，目前讓gin先處理掉大部分的參數
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

// must be main
func main() {
	lambda.Start(Handler)
}
