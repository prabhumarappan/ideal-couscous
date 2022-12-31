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

type Payload struct {
	Data string `json:"data"`
}

type TempData struct {
	DeviceId    int32
	Timestamp   time.Time
	Temperature float64 `json:"Temperature"`
}

type ResponseData struct {
	Overtemp      bool   `json:"overtemp"`
	DeviceId      int32  `json:"device_id,omitempty"`
	FormattedTime string `json:"formatted_time,omitempty"`
}

type ErrorsData struct {
	Errors []string `json:"errors"`
}

var payloadErrors ErrorsData

func verifyPayload(payload Payload) (TempData, error) {
	var tempData TempData
	var err error
	var unixTs int64

	stringSplits := strings.Split(payload.Data, ":")

	if len(stringSplits) != 4 {
		return tempData, errors.New("need 4 parts for the split")
	}

	deviceId, err := strconv.ParseInt(stringSplits[0], 10, 64)
	if err != nil {
		return tempData, err
	}
	tempData.DeviceId = int32(deviceId)

	unixTs, err = strconv.ParseInt(stringSplits[1], 10, 64)
	if err != nil {
		return tempData, err
	}

	tempData.Timestamp = time.Unix(unixTs/1000, unixTs%1000)

	if stringSplits[2] != "'Temperature'" {
		return tempData, errors.New("temperature not found in the payload")
	}

	tempData.Temperature, err = strconv.ParseFloat(stringSplits[3], 64)
	if err != nil {
		return tempData, err
	}

	return tempData, nil
}

func addTemperatureData(c *gin.Context) {
	var requestBody Payload
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		// payloadErrors.Errors = append(payloadErrors.Errors, string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		// payloadErrors.Errors = append(payloadErrors.Errors, string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	tempData, err := verifyPayload(requestBody)
	if err != nil {
		println(err.Error())
		payloadErrors.Errors = append(payloadErrors.Errors, requestBody.Data)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	var responseData ResponseData
	responseData.Overtemp = false
	if tempData.Temperature >= 90 {
		responseData = ResponseData{
			Overtemp:      true,
			DeviceId:      tempData.DeviceId,
			FormattedTime: tempData.Timestamp.Format("2006-01-02 15:04:05")}
	}

	c.JSON(http.StatusOK, responseData)
}

func getErrors(c *gin.Context) {
	c.JSON(http.StatusOK, payloadErrors)
}

func deleteErrors(c *gin.Context) {
	payloadErrors.Errors = nil
	c.JSON(http.StatusNoContent, gin.H{"message": "success"})
}

func main() {
	router := gin.Default()
	router.POST("/temp", addTemperatureData)
	router.GET("/errors", getErrors)
	router.DELETE("/errors", deleteErrors)

	router.Run("0.0.0.0:8080")
}
