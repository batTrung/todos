package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var newTime int64 = 0
var newCount int64 = 0

var normalTime int64 = 0
var normalCount int64 = 0

var poolTime int64 = 0
var poolCount int64 = 0

var dsn = "postgres://db_user:db_user_password@localhost:5433/db_name?sslmode=disable"
var query = "SELECT id, name, price, description FROM products limit 1000"

func scanProducts(rows *sql.Rows) ([]*model.Product, error) {
	defer rows.Close()

	products := make([]*model.Product, 0)
	for rows.Next() {
		var p model.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Description)
		if err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, nil
}

func main() {
	idleConn := 4
	maxConnections := 4
	maxConnLifetime := 2 * time.Minute

	poolConn, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer poolConn.Close()
	poolConn.SetMaxOpenConns(maxConnections)
	poolConn.SetMaxIdleConns(idleConn)
	poolConn.SetConnMaxLifetime(maxConnLifetime)

	conn, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	conn.SetMaxIdleConns(1)

	// Initialize the HTTP router
	router := gin.Default()
	router.StaticFile("/", "./index.html")

	/*
		products/normal: Singleton connection
	*/
	router.GET("/products/normal", func(c *gin.Context) {
		startTime := time.Now()

		// Query the database for all products
		rows, err := conn.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products, err := scanProducts(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		elapsed := time.Since(startTime).Microseconds()
		normalCount++
		normalTime += elapsed
		c.JSON(http.StatusOK, model.Response{Elapsed: elapsed, Average: float64(normalTime / normalCount), Products: products})
	})

	/*
		products/new: New
	*/
	router.GET("/products/new", func(c *gin.Context) {
		startTime := time.Now()
		conn, err := sqlx.Open("postgres", dsn)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}

		rows, err := conn.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products, err := scanProducts(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		elapsed := time.Since(startTime).Microseconds()
		newCount++
		newTime += elapsed
		c.JSON(http.StatusOK, model.Response{Elapsed: elapsed, Average: float64(newTime / newCount), Products: products})
	})

	/*
		products/pooled: Connection Pool
	*/
	router.GET("/products/pooled", func(c *gin.Context) {
		startTime := time.Now()
		// Query the database for all products
		rows, err := poolConn.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products, err := scanProducts(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		elapsed := time.Since(startTime).Microseconds()
		poolCount++
		poolTime += elapsed
		c.JSON(http.StatusOK, model.Response{Elapsed: elapsed, Average: float64(poolTime / poolCount), Products: products})
	})

	// Start the HTTP server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Unable to start HTTP server: %v\n", err)
	}
}
