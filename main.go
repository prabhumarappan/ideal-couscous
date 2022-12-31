package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Payload is the request body
type Payload struct {
	Data string `json:"data"`
}

// TempData is the data from the data string
type TempData struct {
	DeviceId    int32
	Timestamp   time.Time
	Temperature float64 `json:"Temperature"`
}

// ResponseData is the response body that needs to be sent back
type ResponseData struct {
	Overtemp      bool   `json:"overtemp"`
	DeviceId      int32  `json:"device_id,omitempty"`
	FormattedTime string `json:"formatted_time,omitempty"`
}

// ErrorsData is all the data with validation errors that have been sent by the client
type ErrorsData struct {
	Errors []string `json:"errors"`
}

var payloadErrors ErrorsData

// verifyPayload verifies the payload, makes sure it has all the data that is needed and returns the TempData
func verifyPayload(payload Payload) (TempData, error) {
	var tempData TempData
	var err error
	var unixTs int64

	// split the data string into an array by the :
	stringSplits := strings.Split(payload.Data, ":")

	// check if there are 4 parts to the split otherwise return an error
	if len(stringSplits) != 4 {
		return tempData, errors.New("need 4 parts for the split")
	}

	// parse the device id
	deviceId, err := strconv.ParseInt(stringSplits[0], 10, 64)
	if err != nil {
		return tempData, err
	}
	// set the device id
	tempData.DeviceId = int32(deviceId)

	// parse the timestamp
	unixTs, err = strconv.ParseInt(stringSplits[1], 10, 64)
	if err != nil {
		return tempData, err
	}

	// set the timestamp in the time.Time format
	tempData.Timestamp = time.Unix(unixTs/1000, unixTs%1000)

	// check if the temperature string is in the payload otherwise return an error
	if stringSplits[2] != "'Temperature'" {
		return tempData, errors.New("temperature not found in the payload")
	}

	// parse the temperature
	tempData.Temperature, err = strconv.ParseFloat(stringSplits[3], 64)
	if err != nil {
		return tempData, err
	}

	return tempData, nil
}

// addTemperatureData is the handler for the POST /temp endpoint
func addTemperatureData(c *gin.Context) {
	var requestBody Payload
	// read the request body
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		// payloadErrors.Errors = append(payloadErrors.Errors, string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	// unmarshal the request body into the requestBody struct
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		// payloadErrors.Errors = append(payloadErrors.Errors, string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	// verify the payload and return response if there are errors
	tempData, err := verifyPayload(requestBody)
	if err != nil {
		payloadErrors.Errors = append(payloadErrors.Errors, requestBody.Data)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	// build the response data
	var responseData ResponseData
	responseData.Overtemp = false
	if tempData.Temperature >= 90 {
		responseData = ResponseData{
			Overtemp:      true,
			DeviceId:      tempData.DeviceId,
			FormattedTime: tempData.Timestamp.Format("2006-01-02 15:04:05")}
	}

	// send back the response
	c.JSON(http.StatusOK, responseData)
}

// getErrors is the handler for the GET /errors endpoint
// it returns the payloadErrors.Errors as a JSON array
func getErrors(c *gin.Context) {
	c.JSON(http.StatusOK, payloadErrors)
}

// deleteErrors is the handler for the DELETE /errors endpoint
// it sets the payloadErrors.Errors to nil and sends back a 204
func deleteErrors(c *gin.Context) {
	// set the payloadErrors.Errors to nil
	payloadErrors.Errors = nil
	c.JSON(http.StatusNoContent, gin.H{"message": "success"})
}

// main is the entry point for the application
func main() {
	// initialize the gin router
	router := gin.Default()

	// set the routes
	router.POST("/temp", addTemperatureData)
	router.GET("/errors", getErrors)
	router.DELETE("/errors", deleteErrors)

	// start the server
	router.Run("0.0.0.0:8080")
}
