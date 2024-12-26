package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strconv"

    "github.com/spf13/cobra"
)

var city string

func init() {
    rootCmd.PersistentFlags().StringVarP(&city, "city", "c", "", "Name der Stadt für die Wetterabfrage")
    rootCmd.MarkPersistentFlagRequired("city")
}

var rootCmd = &cobra.Command{
    Use:   "app",
    Short: "Eine Anwendung zur Wetterabfrage",
    Run: func(cmd *cobra.Command, args []string) {
        lat, lon, err := getCoordinates(city)
        if err != nil {
            fmt.Printf("Error fetching coordinates: %v\n", err)
            os.Exit(1)
        }

        weatherData, err := fetchWeather(lat, lon)
        if err != nil {
            fmt.Printf("Error fetching weather: %v\n", err)
            os.Exit(1)
        }

        fmt.Printf("Current weather in %s:\n", weatherData.Name)
        fmt.Printf("Temperature: %.2f°C\n", weatherData.Main.Temp)
        if len(weatherData.Weather) > 0 {
            fmt.Printf("Description: %s\n", weatherData.Weather[0].Description)
        }
    },
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func getCoordinates(city string) (float64, float64, error) {
    url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", city)

    resp, err := http.Get(url)
    if err != nil {
        return 0, 0, err
    }
    defer resp.Body.Close()

    var result []struct {
        Lat string `json:"lat"`
        Lon string `json:"lon"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return 0, 0, err
    }

    if len(result) == 0 {
        return 0, 0, fmt.Errorf("no coordinates found for city: %s", city)
    }

    lat, err := strconv.ParseFloat(result[0].Lat, 64)
    if err != nil {
        return 0, 0, err
    }

    lon, err := strconv.ParseFloat(result[0].Lon, 64)
    if err != nil {
        return 0, 0, err
    }

    return lat, lon, nil
}

func fetchWeather(lat, lon float64) (WeatherData, error) {
    url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true", lat, lon)

    resp, err := http.Get(url)
    if err != nil {
        return WeatherData{}, err
    }
    defer resp.Body.Close()

    var weatherData struct {
        CurrentWeather struct {
            Temperature float64 `json:"temperature"`
            Weathercode int     `json:"weathercode"`
        } `json:"current_weather"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
        return WeatherData{}, err
    }

    return WeatherData{
        Name: city,
        Main: MainData{Temp: weatherData.CurrentWeather.Temperature},
        Weather: []WeatherDescription{
            {Description: getWeatherDescription(weatherData.CurrentWeather.Weathercode)},
        },
    }, nil
}

func getWeatherDescription(code int) string {
    switch code {
    case 0:
        return "clear sky"
    case 1, 2, 3:
        return "partly cloudy"
    case 45, 48:
        return "fog"
    case 51, 53, 55:
        return "drizzle"
    case 61, 63, 65:
        return "rain"
    case 71, 73, 75:
        return "snow"
    case 80, 81, 82:
        return "rain showers"
    case 95, 96, 99:
        return "thunderstorm"
    default:
        return "unknown"
    }
}

type WeatherData struct {
    Name    string
    Main    MainData
    Weather []WeatherDescription
}

type MainData struct {
    Temp float64
}

type WeatherDescription struct {
    Description string
}
