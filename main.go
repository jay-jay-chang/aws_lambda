package main

import (
	// go內建
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	// gin框架
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	//redis
	redis "github.com/redis/go-redis/v9"

	//dynamodb
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
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

	//測試redis client
	r.GET("/test_redis", func(c *gin.Context) {
		TestRedisClient()
		c.JSON(200, gin.H{
			"message": "test redis",
		})
	})

	r.GET("/test_db_insert", func(c *gin.Context) {
		ExampleDynamodbInsert()
		c.JSON(200, gin.H{
			"message": "test_db_insert",
		})
	})

	r.GET("/test_db_get", func(c *gin.Context) {
		ExampleDynamodbGet()
		c.JSON(200, gin.H{
			"message": "test_db_get",
		})
	})

	// instance GinLambda object
	ginLambda = ginadapter.New(r)
}

// ---------------------------------------------------------------------
// redis
// ---------------------------------------------------------------------
var redis_ctx = context.Background()

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

//---------------------------------------------------------------------
// dynamoDB
//---------------------------------------------------------------------

func getDynamoDBClient() *dynamodb.DynamoDB {
	// snippet-start:[dynamodb.go.create_new_item.session]
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return dynamodb.New(sess)
	// snippet-end:[dynamodb.go.create_new_item.session]
}

type Item struct {
	Year   int
	Title  string
	Plot   string
	Rating float64
}

func AddTableItem(svc dynamodbiface.DynamoDBAPI, year *int, table, title, plot *string, rating *float64) error {
	// snippet-start:[dynamodb.go.create_new_item.assign_struct]
	item := Item{
		Year:   *year,
		Title:  *title,
		Plot:   *plot,
		Rating: *rating,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	// snippet-end:[dynamodb.go.create_new_item.assign_struct]
	if err != nil {
		return err
	}

	// snippet-start:[dynamodb.go.create_new_item.call]
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: table,
	})
	// snippet-end:[dynamodb.go.create_new_item.call]
	if err != nil {
		return err
	}

	return nil
}

func ExampleDynamodbInsert() {
	// snippet-start:[dynamodb.go.create_new_item.args]
	table := aws.String("test_table")    // "The name of the database table"
	year := aws.Int(2011)                //"The year the movie debuted"
	title := aws.String("test_title_01") // "The title of the movie"
	plot := aws.String("123")            // "The plot of the movie"
	rating := aws.Float64(5.5)           // "The movie rating, from 0.0 to 10.0")

	// snippet-end:[dynamodb.go.create_new_item.args]

	// snippet-start:[dynamodb.go.create_new_item.session]
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)
	// snippet-end:[dynamodb.go.create_new_item.session]

	err := AddTableItem(svc, year, table, title, plot, rating)
	if err != nil {
		fmt.Println("Got an error adding item to table:")
		fmt.Println(err)
		return
	}

	fmt.Println("Successfully added '"+*title+"' ("+strconv.Itoa(*year)+") to table "+*table+" with rating", *rating)
}

func GetTableItem(svc dynamodbiface.DynamoDBAPI, table, title *string, year *int) (*Item, error) {
	// snippet-start:[dynamodb.go.get_item.call]
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: table,
		Key: map[string]*dynamodb.AttributeValue{
			"year": {
				N: aws.String(strconv.Itoa(*year)),
			},
			"title": {
				S: title,
			},
		},
	})
	// snippet-end:[dynamodb.go.get_item.call]
	if err != nil {
		return nil, err
	}

	// snippet-start:[dynamodb.go.get_item.unmarshall]
	if result.Item == nil {
		msg := "Could not find '" + *title + "'"
		return nil, errors.New(msg)
	}

	item := Item{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	// snippet-end:[dynamodb.go.get_item.unmarshall]
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func ExampleDynamodbGet() {
	// snippet-start:[dynamodb.go.get_item.args]
	table := aws.String("test_table")    // "The table to retrieve item from"
	title := aws.String("test_title_01") //"The name of the movie"
	year := aws.Int(2011)                // "The year the movie was released"

	// snippet-end:[dynamodb.go.get_item.args]

	// snippet-start:[dynamodb.go.get_item.session]
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	// snippet-end:[dynamodb.go.get_item.session]

	item, err := GetTableItem(svc, table, title, year)
	if err != nil {
		fmt.Println("Got an error retrieving the item:")
		fmt.Println(err)
		return
	}

	if item == nil {
		fmt.Println("Could not find the table entry")
		return
	}

	fmt.Println("Found item:")
	fmt.Println("Year:  ", item.Year)
	fmt.Println("Title: ", item.Title)
	fmt.Println("Plot:  ", item.Plot)
	fmt.Println("Rating:", item.Rating)
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
