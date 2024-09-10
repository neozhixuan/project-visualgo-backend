package websocketClient

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/neozhixuan/project-visualgo-backend/pb"

	"github.com/gorilla/websocket"
	"github.com/neozhixuan/project-visualgo-backend/data-ingest/utils"
)

func BinanceConnection(broadcast chan []byte, tradeDataChan chan *pb.KlineData) {
	binanceURL := url.URL{
		Scheme: "wss",
		Host:   "stream.binance.com:9443",
		Path:   "/ws/bnbbtc@trade",
	}
	log.Printf("Connecting to Binance WebSocket at %s", binanceURL.String())
	c, _, err := websocket.DefaultDialer.Dial(binanceURL.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect to Binance WebSocket: %v", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// Handle incoming messages from Binance in a goroutine
	go func() {
		defer close(done)

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			// Make sure that broadcast doesn't block indefinitely
			select {
			case broadcast <- message:
				// Successfully sent to broadcast
				log.Println("Sent a message to WSS client")
			default:
				// Handle when no one is reading from broadcast (could log or handle differently)
				log.Println("Warning: broadcast channel is full, dropping message")
			}

			// Unmarshal our raw message into the specified structure
			// - unmarshal = parse the JSON message into our interface
			var msgMap map[string]interface{}
			if err := json.Unmarshal(message, &msgMap); err != nil {
				log.Printf("Error unmarshaling into map: %v", err)
				continue
			}

			// If this is a kline, send data to the gRPC client.
			if eventType, ok := msgMap["e"].(string); ok && eventType == "kline" {
				kData, ok := msgMap["k"].(map[string]interface{})
				if !ok {
					log.Println("Error extracting kline data from message")
					continue
				}

				klineData := &pb.KlineData{
					Symbol:        msgMap["s"].(string),
					OpenTime:      int64(kData["t"].(float64)),
					CloseTime:     int64(kData["T"].(float64)),
					OpenPrice:     utils.ParsePrice(kData["o"]),
					ClosePrice:    utils.ParsePrice(kData["c"]),
					HighPrice:     utils.ParsePrice(kData["h"]),
					LowPrice:      utils.ParsePrice(kData["l"]),
					Volume:        utils.ParsePrice(kData["v"]),
					NumTrades:     int32(kData["n"].(float64)),
					IsKlineClosed: kData["x"].(bool),
				}

				// Non-blocking write to tradeDataChan to avoid deadlock
				//the program is in a deadlock situation where all goroutines are either waiting for something (like receiving or sending on a channel), but no progress can be made because they are waiting indefinitely.
				// - No goroutines can proceed because they are waiting for each other.
				// - Channels are not receiving the data they are expecting or are blocked indefinitely.

				select {
				case tradeDataChan <- klineData:
					// Successfully sent to tradeDataChan
					log.Println("Sent a message to gRPC client")
				default:
					// Handle when tradeDataChan is full (log or take action)
					// the goroutine that reads from the WebSocket connection is trying to send a message through the broadcast channel, but there may not be any receiver actively pulling from this channel, causing the sending goroutine to block indefinitely.
					// The main goroutine is likely waiting for a result from a channel that no other goroutine is writing to, leading to a deadlock.
					log.Println("Warning: tradeDataChan is full, dropping kline data")
				}
			}
		}
	}()

	// Send subscription message to the Binance WSS Connection
	subscribe := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			"bnbbtc@kline_1m", // Stream name
		},
		"id": 1,
	}
	if err := c.WriteJSON(subscribe); err != nil {
		log.Fatal("write:", err)
	}

	// Wait for the goroutine to signal that it has finished
	<-done
}
