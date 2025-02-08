package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// MeterReading represents the hourly meter(consumption) for each day
type MeterReading struct {
	Timestamp int64   `json:"timestamp"` // timestamp in milliseconds
	Reading   float64 `json:"reading"`   // Energy consumption in kwh
}

// MarketPrice represents the hourly price
type MarketPrice struct {
	StartTimestamp int64   `json:"start_timestamp"`
	EndTimestamp   int64   `json:"end_timestamp"`
	MarketPrice    float64 `json:"marketprice"`
	Unit           string  `json:"unit"`
}

// get marketPrices from awattar according the start and end time of the day
func getMarketPrices(start int64, end int64) ([]MarketPrice, error) {
	url := fmt.Sprintf("https://api.awattar.at/v1/marketdata?start=%d&end=%d", start, end)
	response, error := http.Get(url)

	if error != nil {
		return nil, fmt.Errorf("failed to fetch market prices: %v", error)
	}

	defer response.Body.Close()

	//Read the raw response body
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Decode JSON into the expected struct
	var prices struct {
		Data []MarketPrice `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &prices); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	return prices.Data, nil
}

func main() {
	r := gin.Default()

	// To access api at go, cros should be enabled for axios call
	r.Use(cors.Default())

	r.POST("/energy_cost", func(c *gin.Context) {
		var readings []MeterReading

		// Bind incoming request body to the 'reading' variable
		if error := c.BindJSON(&readings); error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Fetch EPEX spot prices for the relevant time of the meter readings
		startTimestamp := readings[0].Timestamp // start timestamp in the first metereReading
		endTimestamp := readings[24].Timestamp  // End timestamp in the last reading
		prices, err := getMarketPrices(startTimestamp, endTimestamp)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch market prices"})
			return
		}

		// Calculate energy consumption in each hour of the day
		var hourlyConsumption [24]float64
		for i := 1; i < len(readings); i++ {
			hourlyConsumption[i-1] = readings[i].Reading - readings[i-1].Reading
		}

		// Multiply energy price of each hour with the respective consumption in that hour.
		var totalCost float64
		for i := 0; i < len(prices); i++ {
			totalCost += hourlyConsumption[i] * (prices[i].MarketPrice / 1000)
		}

		c.JSON(200, gin.H{
			"total_cost": totalCost,
		})

	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
