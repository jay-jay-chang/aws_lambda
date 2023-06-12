package main

import (
	// go內建
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	// package別名
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	// gin框架
	"github.com/gin-gonic/gin"
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

	// instance GinLambda object
	ginLambda = ginadapter.New(r)
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
