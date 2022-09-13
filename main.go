package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/port-labs/tf-connector/port"
	"github.com/port-labs/tf-connector/terraform"
	"go.uber.org/zap"
)

var (
	PORT               string
	TEMPLATES_FOLDER   string
	PORT_CLIENT_ID     string
	PORT_CLIENT_SECRET string

	logger *zap.SugaredLogger
)

func init() {
	flag.StringVar(&PORT, "port", "8080", "Port to listen on")
	flag.StringVar(&TEMPLATES_FOLDER, "templates", "templates", "Folder containing terraform templates")
	PORT_CLIENT_ID, _ = os.LookupEnv("PORT_CLIENT_ID")
	if PORT_CLIENT_ID == "" {
		logger.Fatal("PORT_CLIENT_ID is not set")
	}
	PORT_CLIENT_SECRET, _ = os.LookupEnv("PORT_CLIENT_SECRET")
	if PORT_CLIENT_SECRET == "" {
		logger.Fatal("PORT_CLIENT_SECRET is not set")
	}

	flag.Parse()
}

func main() {
	l, _ := zap.NewProduction()
	defer l.Sync()
	logger := l.Sugar()

	logger.Info("Starting terraform connector")
	logger.Info("Installing terraform on machine")
	tf := terraform.Terraform{}
	err := tf.Install(context.Background())
	if err != nil {
		logger.Fatalf("Failed to install terraform: %v", err)
	}
	r := gin.Default()

	// set metadata for request
	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("templateFolder", TEMPLATES_FOLDER)
		c.Next()
	})

	r.POST("/action", func(c *gin.Context) {
		body := port.ActionBody{}
		err := c.BindJSON(&body)
		if err != nil {
			logger.Errorf("Failed to parse request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		switch body.Payload.Action.Identifier {
		case "CREATE":
			err = tf.Apply(&body, c)
		case "DELETE":
			err = tf.Destroy(&body, c)
		default:
			logger.Errorf("Unknown action: %s", body.Payload.Action.Identifier)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unknown action: %s", body.Payload.Action.Identifier)})
		}
		if err != nil {
			logger.Errorf("Failed to apply terraform: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})

	r.Run(fmt.Sprintf(":%s", PORT))
}
