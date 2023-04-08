package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
)

const (
	port          = 8081
	defaultLimit  = 10
	bucketName    = "postbacks"
	dataDirectory = "data"
)

type Postback struct {
	Method string            `json:"method"`
	URL    string            `json:"url"`
	ID     string            `json:"id"`
	Args   map[string]string `json:"args,omitempty"`
	Body   string            `json:"body,omitempty"`
}

func savePostback(p *Postback, db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		jsonData, err := json.Marshal(p)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(p.ID), jsonData)
	})
}

func getPostbacks(limit int, db *bolt.DB) ([]Postback, error) {
	var postbacks []Postback

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}

		cursor := bucket.Cursor()

		for k, v := cursor.Last(); k != nil && len(postbacks) < limit; k, v = cursor.Prev() {
			var p Postback
			if err := json.Unmarshal(v, &p); err != nil {
				log.Printf("Failed to unmarshal postback: %s", err)
				continue
			}
			postbacks = append(postbacks, p)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(postbacks) == 0 {
		postbacks = []Postback{}
	}

	return postbacks, nil
}

func deletePostback(id string, db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(id))
	})
}

func postbackHandler(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := strconv.FormatInt(time.Now().UnixNano(), 10)
		var body string
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			bodyBytes, err := c.GetRawData()
			if err != nil {
				c.String(http.StatusBadRequest, "Failed to read request body: "+err.Error())
				return
			}
			body = string(bodyBytes)
		}

		argsMap := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				argsMap[key] = values[0]
			}
		}

		p := &Postback{
			Method: c.Request.Method,
			URL:    c.Request.URL.String(),
			ID:     filename,
			Args:   argsMap,
			Body:   body,
		}

		if err := savePostback(p, db); err != nil {
			c.String(http.StatusInternalServerError, "Failed to save postback data: "+err.Error())
			return
		}

		c.String(http.StatusOK, "OK")
	}
}

func getHandler(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.Request.URL.Query().Get("limit")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = defaultLimit
		}
		postbacks, err := getPostbacks(limit, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read postback data: "+err.Error())
			return
		}

		if len(postbacks) == 0 {
			postbacks = []Postback{}
		}

		c.JSON(http.StatusOK, postbacks)
	}
}

func deleteHandler(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postbackID := c.Param("postback_id")
		err := deletePostback(postbackID, db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to delete postback data: "+err.Error())
			return
		}
		c.String(http.StatusOK, "DELETED")
	}
}

func testHandler(c *gin.Context) {
	header := c.DefaultQuery("header", "Location")
	baseURL := "https://go.to-me.local/"
	userID := "b0fa6dd5c8c778795a16bd2aa44df4807d9022e25ae62de342044ec01e2422ad"
	campaignID := "72a4998aaf1e9829b1dd473cd40f964f95c1d2c97e82ea366b73bc245e7d5e73"
	value := c.DefaultQuery("value", baseURL+"?userId="+userID+"&campaignId="+campaignID)

	c.Redirect(http.StatusMovedPermanently, value)
	c.Header(header, value)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func openDB() (*bolt.DB, error) {
	dbPath := filepath.Join(dataDirectory, "postbacks.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	return db, nil
}

func ignoreFavicon() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/favicon.ico" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func main() {
	if err := os.MkdirAll(dataDirectory, 0755); err != nil {
		log.Fatalf("failed to create data directory: %s", err)
	}

	db, err := openDB()
	if err != nil {
		log.Fatalf("failed to open database: %s", err)
	}
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("failed to close database: %s", err)
		}
	}(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(ignoreFavicon())
	r.GET("/get", getHandler(db))
	r.GET("/test-url", testHandler)
	r.GET("/health", healthHandler)
	r.DELETE("/delete/:postback_id", deleteHandler(db))
	r.Any("/:path", postbackHandler(db))
	r.Any("/", postbackHandler(db))

	if err := r.Run(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}
