package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Global variables
const PORT = "8080"
const FORECAST_URI = "https://api.weather.gov/points/"

// Messages
const (
	MissingLatLon         = "Missing lat or lon query parameter"
	InvalidLatLon         = "Invalid lat or lon value"
	NoForecastURL         = "No forecast URL found for location"
	FailedToGetGrid       = "Failed to get gridpoint info"
	FailedToGetForecast   = "Failed to get forecast"
	FailedToParseGrid     = "Failed to parse gridpoint response"
	FailedToParseForecast = "Failed to parse forecast response"
	NoForecastData        = "No forecast data found"
)

// Temperature
const (
	TempHot      = "Hot"
	TempCold     = "Cold"
	TempModerate = "Moderate"
)

// WeatherData holds the weather information
type WeatherData struct {
	Forecast    string `json:"forecast"`    // detailed forecast text
	Temperature string `json:"temperature"` // current temperature
}

// ForecastData holds the forecast information
type ForecastData struct {
	Properties struct {
		Periods []struct {
			Name             string `json:"name"`             // name of the forecast period
			DetailedForecast string `json:"detailedForecast"` // detailed forecast text
			Temperature      int    `json:"temperature"`      // current temperature
			TemperatureUnit  string `json:"temperatureUnit"`  // unit of temperature
		} `json:"periods"`
	} `json:"properties"`
}

// PointsResp holds the response from the points API
type PointsResp struct {
	Properties struct {
		Forecast string `json:"forecast"` // URL to the forecast data
	} `json:"properties"`
}

// ErrorResponse holds the error schema
type ErrorResponse struct {
	Code    int    `json:"code"`    // error code
	Message string `json:"message"` // error message
}

// GetWeather handles the weather requests
// Standard Gorilla Mux router notation is used to handle the requests
// and extract query parameters.
func GetWeather(w http.ResponseWriter, r *http.Request) {
	var errorResponse ErrorResponse

	// Validate query parameters (lat & lon)
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	if latStr == "" || lonStr == "" {
		errorResponse.Code = http.StatusBadRequest
		errorResponse.Message = MissingLatLon
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Parse latitude and longitude data as decimal degree floating point (not EWLL or DMS)
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if err1 != nil || err2 != nil {
		errorResponse.Code = http.StatusBadRequest
		errorResponse.Message = InvalidLatLon
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Step 1: Get gridpoint info
	pointsURL := fmt.Sprintf("%s/%f,%f", FORECAST_URI, lat, lon)
	resp, err := http.Get(pointsURL)
	if err != nil || resp.StatusCode != 200 {
		errorResponse.Code = http.StatusInternalServerError
		errorResponse.Message = FailedToGetGrid
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	defer resp.Body.Close()

	// Parse the gridpoint response
	var pointsResp PointsResp
	if err := json.NewDecoder(resp.Body).Decode(&pointsResp); err != nil {
		errorResponse.Code = http.StatusInternalServerError
		errorResponse.Message = FailedToParseGrid
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Step 2: Get forecast
	forecastURL := pointsResp.Properties.Forecast
	if forecastURL == "" {
		errorResponse.Code = http.StatusNotFound
		errorResponse.Message = NoForecastURL
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	// Get the forecast response from the forecast URL
	forecastResp, err := http.Get(forecastURL)
	if err != nil || forecastResp.StatusCode != 200 {
		errorResponse.Code = http.StatusInternalServerError
		errorResponse.Message = FailedToGetForecast
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	defer forecastResp.Body.Close()

	// Step 3: Parse forecast data
	var forecastData ForecastData
	if err := json.NewDecoder(forecastResp.Body).Decode(&forecastData); err != nil {
		errorResponse.Code = http.StatusInternalServerError
		errorResponse.Message = FailedToParseForecast
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	if len(forecastData.Properties.Periods) == 0 {
		errorResponse.Code = http.StatusNotFound
		errorResponse.Message = NoForecastData
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	tempUnit := "F"
	// Get the first period's weather data
	// Rewriting TemperatureUnit to match the new temperature scale based on the
	// assignment (Hot, Cold, Moderate).  Not as easy to read coding wise
	// but eliminates making another member variable in the structure for this assignment.
	period := forecastData.Properties.Periods[0]
	if period.TemperatureUnit == "F" {
		if period.Temperature >= 81 {
			period.TemperatureUnit = TempHot // Hot if temperature is 81F or above
		} else if period.Temperature <= 45 {
			period.TemperatureUnit = TempCold // Cold if temperature is 45F or below
		} else {
			period.TemperatureUnit = TempModerate // Moderate if temperature is between 45F and 81F
		}
	} else {
		tempUnit = "C"
		if period.Temperature >= 27 {
			period.TemperatureUnit = TempHot // Hot if temperature is 27C or above
		} else if period.Temperature <= 7 {
			period.TemperatureUnit = TempCold // Cold if temperature is 7C or below
		} else {
			period.TemperatureUnit = TempModerate // Moderate if temperature is between 7C and 27C
		}

	}
	// Determine temperature description
	weather := WeatherData{
		Forecast:    period.DetailedForecast,
		Temperature: fmt.Sprintf("%s (%d %s)", period.TemperatureUnit, period.Temperature, tempUnit),
	}
	if len(forecastData.Properties.Periods) == 0 {
		errorResponse.Code = http.StatusNotFound
		errorResponse.Message = NoForecastData
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Return the weather data as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weather)
}

func main() {
	r := mux.NewRouter()
	// Would normally build a routers file with all the
	// routes listed, but since it's just 1 route, will
	// just handle the function here
	r.HandleFunc("/GetWeather", GetWeather).Methods("GET")
	// Hardcode the port for now
	// Would add a configuration file or environment variable later
	port := PORT
	// Could do something like this for port variable
	/*port := os.Getenv("PORT")
	if port == "" {
		port = PORT
	}*/
	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
