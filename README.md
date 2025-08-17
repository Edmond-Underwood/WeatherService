# WeatherService - A coding assignment from TekSystems (Jack Henry)

This project is a Go HTTP server that exposes a `GetWeather` endpoint. The endpoint accepts decimal latitude and longitude as query parameters, calls the National Weather Service API (https://api.weather.gov/), and returns a `WeatherData` struct with forecast and temperature description fields.

## Features
- HTTP server using gorilla-mux
- `/GetWeather` endpoint
- Accepts decimal degree `lat` and `lon` as query parameters
- Calls the National Weather Service API
- Returns weather forecast and temperature description
- All code in 1 file, main.go for assignment purposes and ease of development and dev building

## Requirements
- Go 1.18+
- gorilla/mux

## Running the server
1. Install dependencies:
   ```sh
   go mod tidy
   ```
2. Run the server:
   ```sh
   go run main.go
   ```

## Example request
```
GET http://localhost:8080/GetWeather?lat=38.8977&lon=-77.0365
e.g. curl GET "http://localhost:8080/GetWeather?lat=38&lon=-97"
```

## License - For TekSystems and Clients Dev Usage
MIT
