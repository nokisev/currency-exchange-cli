package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ExchangeRates struct {
	Result             string             `json:"result"`
	Documentation      string             `json:"documentation"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	BaseCode           string             `json:"base_code"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
}

func getExchangeRates(apiKey, baseCurrency string) (*ExchangeRates, error) {
	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", apiKey, baseCurrency)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	var rates ExchangeRates
	if err := json.Unmarshal(body, &rates); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	if rates.Result != "success" {
		return nil, fmt.Errorf("API вернул ошибку: %s", rates.Result)
	}

	return &rates, nil
}

func convertCurrency(amount float64, from, to string, rates *ExchangeRates) (float64, error) {
	fromRate, ok := rates.ConversionRates[from]
	if !ok {
		return 0, fmt.Errorf("валюта %s не найдена", from)
	}

	toRate, ok := rates.ConversionRates[to]
	if !ok {
		return 0, fmt.Errorf("валюта %s не найдена", to)
	}

	return (amount / fromRate) * toRate, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println(".env не существует")
		return
	}
	apiKey := os.Getenv("API_KEY")
	baseCurrency := "USD"

	rates, err := getExchangeRates(apiKey, baseCurrency)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	amount, _ := strconv.ParseFloat(os.Args[1], 64)
	from := os.Args[2]
	to := os.Args[3]

	converted, err := convertCurrency(amount, from, to, rates)
	if err != nil {
		fmt.Printf("Ошибка конвертации: %v\n", err)
		return
	}
	fmt.Printf("%.2f %s = %.2f %s\n", amount, from, converted, to)
}
