package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// WeatherResponse represents the structure of the weather API response
type WeatherResponse struct {
	Name string `json:"name"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

func fetchWeather(city string) (WeatherResponse, error) {
	// Beispiel: Feste Koordinaten für Berlin
	latitude := "52.5200"
	longitude := "13.4050"

	var weatherData struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			WeatherCode int     `json:"weathercode"`
		} `json:"current_weather"`
	}

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current_weather=true", latitude, longitude)
	resp, err := http.Get(url)
	if err != nil {
		return WeatherResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherResponse{}, fmt.Errorf("failed to fetch weather data: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&weatherData)
	if err != nil {
		return WeatherResponse{}, err
	}

	return WeatherResponse{
		Name: "Berlin",
		Main: struct {
			Temp float64 `json:"temp"`
		}{
			Temp: weatherData.CurrentWeather.Temperature,
		},
		Weather: []struct {
			Description string `json:"description"`
		}{
			{Description: fmt.Sprintf("Code %d", weatherData.CurrentWeather.WeatherCode)},
		},
	}, nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "weathercli",
		Short: "Weather CLI tool",
		Long:  "A simple CLI tool to fetch and display the current weather using OpenWeatherMap API.",
	}

	var printCmd = &cobra.Command{
		Use:   "print",
		Short: "Print the current weather",
		Run: func(cmd *cobra.Command, args []string) {
			city := viper.GetString("city")
			apiKey := viper.GetString("apikey")

			if apiKey == "" {
				fmt.Println("Error: API key is required. Set it in the config file or pass it as an environment variable.")
				os.Exit(1)
			}

			if city == "" {
				fmt.Println("Error: City is required. Set it in the config file or pass it as an argument.")
				os.Exit(1)
			}

			weatherData, err := fetchWeather(city)
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

	rootCmd.AddCommand(printCmd)

	// Use Viper for configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Set default values
	viper.SetDefault("city", "Berlin")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Warning: No configuration file found. Using defaults.")
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
