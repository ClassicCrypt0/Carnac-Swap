package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Kucoin/kucoin-go-sdk"
)

const (
	KUCOIN_API_URL       = "https://api.kucoin.com/api/v1/market/orderbook/level1"
	FILE_NAME            = "trades.csv"
	baseValidationCheck  = .9995
	quoteValidationCheck = 1.0005
	INITIAL_AMOUNT       = 100
)

var (
	SELL_THRESHOLD     float64
	BASE_AMOUNT        string
	QUOTE_AMOUNT       string
	CUSTOM_PAIRS       string
	TELEGRAM_BOT_TOKEN string
	TELEGRAM_CHAT_ID   string
)

type PaperAccount map[string]float64
type LastThresholdPrices map[string]float64

type Data struct {
	Price string `json:"price"`
}

type CreateOrderResponse struct {
	OrderId string `json:"orderId"`
}

type ResponseData struct {
	Data struct {
		Time        int64  `json:"time"`
		Sequence    string `json:"sequence"`
		Price       string `json:"price"`
		Size        string `json:"size"`
		BestBid     string `json:"bestBid"`
		BestBidSize string `json:"bestBidSize"`
		BestAsk     string `json:"bestAsk"`
		BestAskSize string `json:"bestAskSize"`
	} `json:"data"`
}

type Config struct {
	SellThreshold    float64 `json:"sell_threshold"`
	BaseAmount       string  `json:"base_amount"`
	QuoteAmount      string  `json:"quote_amount"`
	CustomPairs      string  `json:"custom_pairs"`
	TelegramBotToken string  `json:"telegram_bot_token"`
	TelegramChatID   string  `json:"telegram_chat_id"`
	ApiKey           string  `json:"api_key"`
	SecretKey        string  `json:"secret_key"`
	Passphrase       string  `json:"passphrase"`
}

func main() {
	cfg := readConfig()
	SELL_THRESHOLD = cfg.SellThreshold
	BASE_AMOUNT = cfg.BaseAmount
	QUOTE_AMOUNT = cfg.QuoteAmount
	CUSTOM_PAIRS = cfg.CustomPairs
	TELEGRAM_BOT_TOKEN = cfg.TelegramBotToken
	TELEGRAM_CHAT_ID = cfg.TelegramChatID

	customPairs := strings.Split(CUSTOM_PAIRS, ",")
	monitorCustomPairs(customPairs)
}

func readConfig() Config {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	json.Unmarshal(bytes, &cfg)
	return cfg
}

func sendMessageToTelegram(message string) {
	telegramApi := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TELEGRAM_BOT_TOKEN)
	params := fmt.Sprintf("?chat_id=%s&text=%s", TELEGRAM_CHAT_ID, message)
	resp, err := http.Get(telegramApi + params)

	if err != nil {
		fmt.Printf("Error sending message to Telegram: %v\n", err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Error reading Telegram response: %v\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Error sending message to Telegram: %v\n", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if !result["ok"].(bool) {
		fmt.Printf("Telegram API Error: %v\n", result["description"].(string))
		return
	}
}

func executeLiveTrade(action string, side string, symbol string, orderType string, price string, size string) {
	// Load the configuration file
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config file: ", err)
		return
	}
	defer configFile.Close()

	// Decode the JSON file into the Config struct
	config := Config{}
	bytes, _ := ioutil.ReadAll(configFile)
	if err := json.Unmarshal(bytes, &config); err != nil {
		fmt.Println("Error decoding config JSON: ", err)
		return
	}

	// Use the credentials from the config file
	s := kucoin.NewApiService(
		kucoin.ApiKeyOption(config.ApiKey),
		kucoin.ApiSecretOption(config.SecretKey),
		kucoin.ApiPassPhraseOption(config.Passphrase),
		kucoin.ApiKeyVersionOption(kucoin.ApiKeyVersionV2),
	)

	switch action {
	case "accounts":
		resp, err := s.Accounts("", "")
		if err != nil {
			fmt.Println(err)
			return
		}

		as := kucoin.AccountsModel{}
		if err := resp.ReadData(&as); err != nil {
			fmt.Println(err)
			return
		}

		for _, a := range as {
			if a.Type == "trade" {
				fmt.Printf("%s: %s %s\n", a.Id, a.Currency, a.Balance)
			}
		}
	case "place":
		baseCurrency := strings.Split(symbol, "-")[0]

		// get balance for the base currency of the trading pair
		balanceResp, err := s.Accounts(baseCurrency, "")
		if err != nil {
			fmt.Println(err)
			return
		}
		balanceModel := kucoin.AccountsModel{}
		if err := balanceResp.ReadData(&balanceModel); err != nil {
			fmt.Println(err)
			return
		}

		var tradingAccount kucoin.AccountModel
		for _, a := range balanceModel {
			if a.Type == "trade" {
				tradingAccount = *a
				break
			}
		}

		if tradingAccount.Id == "" {
			fmt.Println("No trading account found for the given symbol.")
			return
		}

		balance, err := strconv.ParseFloat(tradingAccount.Balance, 64)
		if err != nil {
			fmt.Println("Invalid balance: ", err)
			return
		}

		// Calculate order size
		var orderSize float64
		if strings.HasSuffix(size, "%") {
			percentage, err := strconv.ParseFloat(strings.TrimSuffix(size, "%"), 64)
			if err != nil {
				fmt.Println("Invalid percentage: ", err)
				return
			}
			orderSize = balance * percentage / 100
		} else {
			orderSize, err = strconv.ParseFloat(size, 64)
			if err != nil {
				fmt.Println("Invalid size: ", err)
				return
			}
		}

		// Round order size to 5 decimal places
		roundedSize := math.Round(orderSize*1e5) / 1e5
		size = strconv.FormatFloat(roundedSize, 'f', 4, 64)

		rsp, err := s.CreateOrder(&kucoin.CreateOrderModel{
			Side:      side,
			Symbol:    symbol,
			Type:      orderType,
			Price:     price,
			Size:      size,
			ClientOid: "my-unique-order-id",
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		orderResponse := CreateOrderResponse{}
		if err := rsp.ReadData(&orderResponse); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Order placed. ID: %s\n", orderResponse.OrderId)
	}

	// Sending a message to the telegram channel
	sendMessageToTelegram(url.QueryEscape(fmt.Sprintf("New trade executed:\n%s\n", strings.Join([]string{action, side, symbol, size}, " "))))
}

func monitorCustomPairs(customPairs []string) {
	// Sending a message to the telegram channel
	sendMessageToTelegram(fmt.Sprintf("Carnac Swap Started"))
	initialPrices := make(map[string]float64)
	lastThresholdPrices := getLastThresholdPrices()

	paperAccount := createPaperAccount(customPairs, INITIAL_AMOUNT)

	for _, pair := range customPairs {
		baseSymbol, quoteSymbol := getBaseAndQuoteSymbol(pair)

		basePrice, _, _ := fetchPrice(baseSymbol)
		quotePrice, _, _ := fetchPrice(quoteSymbol)

		if basePrice == 0 || quotePrice == 0 {
			fmt.Printf("Error fetching initial prices for %v. Skipping.\n", pair)
			continue
		}

		initialPairPrice := basePrice / quotePrice

		if _, ok := lastThresholdPrices[pair]; !ok {
			lastThresholdPrices[pair] = initialPairPrice
		}

		initialPrices[pair] = initialPairPrice
		fmt.Printf("Initial price for %v: %v\n", pair, initialPairPrice)
	}

	for {
		for _, pair := range customPairs {
			baseSymbol, quoteSymbol := getBaseAndQuoteSymbol(pair)

			basePrice, baseBestBid, baseBestAsk := fetchPrice(baseSymbol)
			quotePrice, quoteBestBid, quoteBestAsk := fetchPrice(quoteSymbol)

			if basePrice == 0 || quotePrice == 0 {
				fmt.Printf("Error fetching prices for %v. Skipping.\n", pair)
				continue
			}

			currentPairPrice := basePrice / quotePrice
			lastThresholdPrice := lastThresholdPrices[pair]

			if lastThresholdPrice == 0 {
				// Skip if lastThresholdPrice is zero to avoid division by zero
				continue
			}

			priceChangePercentage := (currentPairPrice - lastThresholdPrice) / lastThresholdPrice * 100

			currentTime := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("[%v] %v basePrice: %.4f, baseBestBid: %.4f, baseBestAsk: %.4f, quotePrice: %.4f, quoteBestBid: %.4f, quoteBestAsk: %.4f\n", currentTime, pair, basePrice, baseBestBid, baseBestAsk, quotePrice, quoteBestBid, quoteBestAsk)
			fmt.Printf("[%v] %v price change percentage: %.2f%%\n", currentTime, pair, priceChangePercentage)

			baseBidAskChangePercentage := ((baseBestBid / quoteBestAsk) - lastThresholdPrice) / lastThresholdPrice * 100
			quoteBidAskChangePercentage := ((baseBestAsk / quoteBestBid) - lastThresholdPrice) / lastThresholdPrice * 100
			fmt.Printf("[%v] %v Base bid/ask change percentage: %.2f%%\n", currentTime, pair, baseBidAskChangePercentage)
			fmt.Printf("[%v] %v Quote bid/ask change percentage: %.2f%%\n", currentTime, pair, quoteBidAskChangePercentage)

			if priceChangePercentage >= SELL_THRESHOLD {
				if baseBidAskChangePercentage >= SELL_THRESHOLD {
					fmt.Printf("Sell base asset in %v\n", pair)
					sellAmount := paperAccount[baseSymbol] * 0.1
					paperAccount = simulateTrade(paperAccount, baseSymbol, sellAmount, basePrice)
					executeLiveTrade("place", "sell", baseSymbol, "market", "1", BASE_AMOUNT)
					time.Sleep(4 * time.Second)

					fmt.Println("Buy quote asset with USDT")
					buyAmount := paperAccount["USDT"] / quotePrice * 0.999 // Account for 0.1% trading fee
					paperAccount[quoteSymbol] += buyAmount
					paperAccount["USDT"] -= buyAmount * quotePrice
					executeLiveTrade("place", "buy", quoteSymbol, "market", "1", "100%")
					time.Sleep(4 * time.Second)

					lastThresholdPrices[pair] = currentPairPrice
					writeToCsv(currentTime, pair, currentPairPrice, priceChangePercentage)
				}
			} else if priceChangePercentage <= -SELL_THRESHOLD {
				if quoteBidAskChangePercentage <= -SELL_THRESHOLD {
					fmt.Printf("Sell quote asset in %v\n", pair)
					sellAmount := paperAccount[quoteSymbol] * 0.1
					paperAccount = simulateTrade(paperAccount, quoteSymbol, sellAmount, quotePrice)
					executeLiveTrade("place", "sell", quoteSymbol, "market", "1", QUOTE_AMOUNT)
					time.Sleep(4 * time.Second)

					fmt.Println("Buy base asset with USDT")
					buyAmount := paperAccount["USDT"] / basePrice * 0.999 // Account for 0.1% trading fee
					paperAccount[baseSymbol] += buyAmount
					paperAccount["USDT"] -= buyAmount * basePrice
					executeLiveTrade("place", "buy", baseSymbol, "market", strconv.FormatFloat(baseBestAsk, 'f', -1, 64), "100%")
					time.Sleep(4 * time.Second)

					lastThresholdPrices[pair] = currentPairPrice
					writeToCsv(currentTime, pair, currentPairPrice, priceChangePercentage)
				}
			}
		}

		//fmt.Printf("Current paper account balance: %v\n", paperAccount)
		//fmt.Printf("Total portfolio value in USDT: %.2f\n", calculatePortfolioValue(paperAccount))
		time.Sleep(5 * time.Second) // Adjust the sleep time as needed (in seconds)
	}
}

func getBaseAndQuoteSymbol(pair string) (string, string) {
	splitPair := strings.Split(pair, "/")
	return splitPair[0], splitPair[1]
}

func fetchPrice(symbol string) (float64, float64, float64) {
	params := symbol
	resp, err := http.Get(KUCOIN_API_URL + "?symbol=" + params)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 0, 0, 0
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 0, 0, 0
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Error: %v\n", resp.StatusCode)
		return 0, 0, 0
	}

	data := &ResponseData{}
	json.Unmarshal(body, &data)

	price, err := strconv.ParseFloat(data.Data.Price, 64)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 0, 0, 0
	}

	bestBid, err := strconv.ParseFloat(data.Data.BestBid, 64)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 0, 0, 0
	}

	bestAsk, err := strconv.ParseFloat(data.Data.BestAsk, 64)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 0, 0, 0
	}

	return price, bestBid, bestAsk
}

func writeToCsv(timestamp, customPair string, price, changePercentage float64) {
	file, err := os.OpenFile("trades.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(file)

	fileStat, _ := file.Stat()
	if fileStat.Size() == 0 {
		csvwriter.Write([]string{"timestamp", "custom pair", "price", "change_percentage"})
	}

	data := []string{timestamp, customPair, strconv.FormatFloat(price, 'f', 6, 64), strconv.FormatFloat(changePercentage, 'f', 2, 64)}
	if err := csvwriter.Write(data); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}

	csvwriter.Flush()
	file.Close()
}

func getLastThresholdPrices() LastThresholdPrices {
	lastPrices := make(LastThresholdPrices)

	file, err := os.Open(FILE_NAME)
	if err != nil {
		return lastPrices
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return lastPrices
	}

	for _, record := range records[1:] {
		customPair, price := record[1], record[2]
		floatPrice, err := strconv.ParseFloat(price, 64)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		lastPrices[customPair] = floatPrice
	}
	return lastPrices
}

func createPaperAccount(customPairs []string, initialAmount float64) PaperAccount {
	paperAccount := make(PaperAccount)
	paperAccount["USDT"] = 0

	file, err := os.Open(FILE_NAME)
	if err != nil || isEmpty(file) {
		for _, pair := range customPairs {
			baseSymbol, quoteSymbol := getBaseAndQuoteSymbol(pair)

			for _, symbol := range []string{baseSymbol, quoteSymbol} {
				var price float64
				if symbol == baseSymbol {
					price, _, _ = fetchPrice(symbol)
				} else if symbol == quoteSymbol {
					_, _, price = fetchPrice(symbol)
				}

				if _, ok := paperAccount[symbol]; !ok {
					paperAccount[symbol] = initialAmount / price
				}
			}

			basePrice, _, _ := fetchPrice(baseSymbol)
			quotePrice, _, _ := fetchPrice(quoteSymbol)
			initialPairPrice := basePrice / quotePrice

			currentTime := time.Now().Format("2006-01-02 15:04:05")
			writeToCsv(currentTime, pair, initialPairPrice, 0.0) // Assuming the change_percentage is 0 for the initial transactions
		}
	}
	file.Close()
	return paperAccount
}

// Check if a file is empty
func isEmpty(f *os.File) bool {
	fileStat, err := f.Stat()
	if err != nil {
		return false
	}
	return fileStat.Size() == 0
}

func simulateTrade(paperAccount PaperAccount, symbol string, amount, price float64) PaperAccount {
	usdtValue := amount * price
	paperAccount[symbol] -= amount
	paperAccount["USDT"] += usdtValue * 0.999 // Account for 0.1% trading fee
	return paperAccount
}

func calculatePortfolioValue(paperAccount PaperAccount) float64 {
	totalValue := 0.0
	for symbol, amount := range paperAccount {
		if symbol != "USDT" {
			price, _, _ := fetchPrice(symbol)
			usdtValue := amount * price
			totalValue += usdtValue
		} else {
			totalValue += amount
		}
	}
	return totalValue
}

// USAGE
// execute the script | action | side | pair | order type | price (limit only) | size %
// go run Carnac_Max.go place sell LUNC-USDT market 1 10
// go run Carnac_Max.go place buy USTC-USDT market 1 100
