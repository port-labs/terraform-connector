package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
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
	PORT_BASE_URL      string

	logger     *zap.SugaredLogger
	portClient *port.Client
	tf         *terraform.Terraform
)

func init() {
	if debug, ok := os.LookupEnv("DEBUG"); ok && debug == "true" {
		l, _ := zap.NewDevelopment()
		defer l.Sync()
		logger = l.Sugar()
	} else {
		l, _ := zap.NewProduction()
		defer l.Sync()
		logger = l.Sugar()
	}

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
	PORT_BASE_URL, _ = os.LookupEnv("PORT_BASE_URL")
	if PORT_BASE_URL == "" {
		PORT_BASE_URL = "https://api.getport.io"
	}

	flag.Parse()

	logger.Info("Starting terraform connector")
	logger.Info("Installing terraform on machine")
	tf = terraform.NewTerraform(logger)
	err := tf.Install(context.Background())
	if err != nil {
		logger.Fatalf("Failed to install terraform: %v", err)
	}
	portClient = port.New(PORT_BASE_URL)
	logger.Info("Authenticating with Port")
	_, err = portClient.Authenticate(context.Background(), PORT_CLIENT_ID, PORT_CLIENT_SECRET)
	if err != nil {
		logger.Fatalf("failed to authenticate with Port: %v", err)
	}
}

func main() {

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		err := verifyHmac(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Next()
	})

	// set metadata for request
	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("templateFolder", TEMPLATES_FOLDER)
		c.Next()
	})

	r.POST("/", func(c *gin.Context) {
		err := actionHandler(c)
		if err != nil {
			logger.Errorf("%v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})

	r.Run(fmt.Sprintf(":%s", PORT))
}

func actionHandler(c *gin.Context) (err error) {
	body := port.ActionBody{}
	err = c.BindJSON(&body)
	if err != nil {
		return err
	}
	switch body.Payload.Action.Trigger {
	case "CREATE", "DAY-2":
		err = tf.Apply(&body, c)
	case "DELETE":
		err = tf.Destroy(&body, c)
	default:
		return fmt.Errorf("unknown action: %s", body.Payload.Action.Identifier)
	}
	if err != nil {
		portClient.PatchActionRun(c, body.Context.RunID, port.ActionStatusFailure)
		return err
	}
	err = portClient.PatchActionRun(c, body.Context.RunID, port.ActionStatusSuccess)
	if err != nil {
		return err
	}
	return nil
}

func verifyHmac(c *gin.Context) (err error) {
	signature := c.GetHeader("x-port-signature")
	timestamp := c.GetHeader("x-port-timestamp")
	if signature == "" {
		return fmt.Errorf("missing x-port-signature")
	}
	// trim 'v1,...'
	signature = signature[3:]
	body, err := c.GetRawData()
	if err != nil {
		return err
	}
	// let other consumers read from body
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	fmt.Println(string(body))
	sign := hmac.New(sha256.New, []byte(PORT_CLIENT_SECRET))
	sign.Write([]byte(fmt.Sprintf("%s.%s", timestamp, string(body))))
	expected := base64.StdEncoding.EncodeToString(sign.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) == 0 {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
