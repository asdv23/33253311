package binanceapi

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type FuturesAPI struct {
	client *futures.Client
}

func NewFuturesAPI(apiKey, secretKey string) *FuturesAPI {
	futures.UseTestnet = true

	client := futures.NewClient(apiKey, secretKey)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.NewPingService().Do(ctx); err != nil {
		panic(err)
	}
	return &FuturesAPI{
		client: futures.NewClient(apiKey, secretKey),
	}
}

func (fapi *FuturesAPI) FetchTrades(c *gin.Context) {
	tickers, err := fapi.client.NewExchangeInfoService().Do(c)
	if err != nil {
		log.Printf("failed to get exchange info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trades"})
		return
	}
	log.Println(len(tickers.Symbols)) // console.log(res.symbols.length)

	tradeStream, err := fapi.userTrades(c, tickers.Symbols, futures.WithRecvWindow(5000))
	if err != nil {
		log.Printf("failed to get user trades: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trades"})
		return
	}

	allTrades := make([]int, 0, len(tradeStream))
	for _, trade := range tradeStream {
		log.Println(len(trade.trades), trade.symbol, len(allTrades)) // console.log(trades.length, symbolInfo.symbol, allTrades.length);
		allTrades = append(allTrades, len(trade.trades))             // allTrades.push(trades.length);
	}

	c.JSON(http.StatusOK, allTrades)
}

func (fapi *FuturesAPI) BTC(c *gin.Context) {
	symbol := "BTCUSDT"
	trades, err := fapi.client.NewRecentTradesService().Symbol(symbol).Do(c)
	if err != nil {
		log.Printf("failed to get trades: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch BTCUSDT futures trades"})
		return
	}

	c.JSON(http.StatusOK, trades)
}

func (fapi *FuturesAPI) Allfut(c *gin.Context) {
	exchangeInfo, err := fapi.client.NewExchangeInfoService().Do(c)
	if err != nil {
		log.Printf("failed to get exchange info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch all futures trades"})
		return
	}

	allTrades, err := fapi.trades(c, exchangeInfo.Symbols, 50 /*limit*/)
	if err != nil {
		log.Printf("failed to get trades: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch all futures trades"})
		return
	}

	c.JSON(http.StatusOK, allTrades)
}

type userTradesBySymbol struct {
	symbol string
	trades []*futures.AccountTrade
}

func (fapi *FuturesAPI) userTrades(c *gin.Context, symbols []futures.Symbol, opts ...futures.RequestOption) ([]*userTradesBySymbol, error) {
	allTradeStream := make(chan *userTradesBySymbol)

	var g errgroup.Group
	for _, symbolInfo := range symbols {
		symbol := symbolInfo.Symbol
		g.Go(func() error {
			trades, err := fapi.client.NewListAccountTradeService().Symbol(symbol).Do(c, opts...)
			if err != nil {
				return err
			}

			allTradeStream <- &userTradesBySymbol{symbol, trades}
			return nil
		})
	}

	go func() {
		g.Wait()
		close(allTradeStream)
	}()

	allTrades := make([]*userTradesBySymbol, 0, len(symbols))
	for trade := range allTradeStream {
		allTrades = append(allTrades, trade)
	}

	return allTrades, g.Wait()
}

type tradesBySymbol struct {
	Symbol string
	Trades []*futures.Trade
}

func (fapi *FuturesAPI) trades(c *gin.Context, symbols []futures.Symbol, limit int, opts ...futures.RequestOption) ([]*tradesBySymbol, error) {
	allTradeStream := make(chan *tradesBySymbol)

	var g errgroup.Group
	for _, symbolInfo := range symbols {
		symbol := symbolInfo.Symbol
		g.Go(func() error {
			trades, err := fapi.client.NewRecentTradesService().Symbol(symbol).Limit(limit).Do(c, opts...)
			if err != nil {
				return err
			}

			allTradeStream <- &tradesBySymbol{symbol, trades}
			return nil
		})
	}

	go func() {
		g.Wait()
		close(allTradeStream)
	}()

	allTrades := make([]*tradesBySymbol, 0, len(symbols))
	for trade := range allTradeStream {
		allTrades = append(allTrades, trade)
	}

	return allTrades, g.Wait()
}
