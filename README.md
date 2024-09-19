# VisuALGO

This project consists of two main services: data-ingest and trading-algo.
The 'data-ingest' service fetches market data from Binance WebSocket,
sends it to clients via both gRPC and WebSocket servers. The 'trading-algo'
service processes the data from gRPC, performs trading calculations like EMA
and VWAP, and sends the results to a WebSocket server.

## Project Structure:

```
├── proto/ # Contains .proto files for defining gRPC services
├── data-ingest/ # Fetches data from Binance, provides gRPC and WebSocket servers
├── trading-algo/ # Performs trading calculations (EMA, VWAP) using gRPC data, sends results via WebSocket
└── README.md # Project documentation
```

## Prerequisites:

- Go (Golang) installed
- `protoc` compiler installed for generating gRPC code
- Make sure you have Go packages for gRPC and WebSockets:
  - `go install google.golang.org/grpc`
  - `go install github.com/gorilla/websocket`
- Binance API keys if needed for WebSocket access

## Setup with Compose

Create .env.local file in /frontend

```bash
NEXT_PUBLIC_STAGE=production
```

## Proto (protobuf folder)

The `proto/` folder contains `.proto` files that define the gRPC services and messages.
The `data-ingest` service acts as a gRPC server, while the `trading-algo` service is a gRPC client.

To generate Go files from the proto definitions, run the following command:

```bash
protoc -I proto/ 
proto/product_info.proto 
--go_out=pb/ --go-grpc_out=pb/
```

## Data-Ingest Service (data-ingest folder)

The `data-ingest` service connects to the Binance WebSocket stream and ingests real-time market data.
It exposes two services:

- gRPC server on port `50051` to send candlestick data to clients.
- WebSocket server on port `8080` that allows clients to receive real-time data via WebSocket.

Start the gRPC and WebSocket server from the `data-ingest` folder:

```bash
go run main.go
```

The `data-ingest` service:

1. Connects to the Binance WebSocket stream to ingest real-time market data (e.g., candlesticks).
2. Sends this data to clients connected via gRPC on port 50051.
3. Sends the same data via a WebSocket server running on port 8080.

## Trading-Algorithm Service (trading-algo folder)

The `trading-algo` service performs trading calculations such as EMA (Exponential Moving Average) and VWAP (Volume Weighted Average Price).
This service connects to the gRPC server running on port `50051` to receive candlestick data.
After performing calculations, it broadcasts the results via a WebSocket server running on port `8090`.

Start the Trading Algorithm Service:

```bash
go run main.go
```

The `trading-algo` service:

1. Receives candlestick data from the gRPC server (port 50051).
2. Calculates trading indicators (EMA and VWAP).
3. Sends the results via WebSocket to clients connected on port 8090.

## Trading Calculations - EMA and VWAP

### EMA (Exponential Moving Average):

EMA is a type of moving average that gives more weight to recent prices, making it more responsive to new information.
In the context of this project, EMA is used to smooth out price data and to identify trends.
The 9-period EMA is commonly used for short-term trend analysis in strategies like the rubberband strategy.

### VWAP (Volume Weighted Average Price):

VWAP is the average price of an asset, weighted by volume. It provides an indication of the true average price over a given period.
VWAP is useful for assessing the "fair" price of an asset during the day.

In the **rubberband price strategy**, both EMA and VWAP are used to assess how far the current price is from its average.

- If the price moves significantly away from the EMA (creating a "rubberband" effect), it often means that the price could bounce back toward the average.
- VWAP helps confirm if the price is above or below its fair value, which can be a signal to take trades based on market behavior.

Example:

- When the price is stretched far above the 9-EMA but remains below VWAP, it might signal overbought conditions and a possible mean reversion trade.
- Conversely, if the price is below the 9-EMA and VWAP, it might indicate oversold conditions.

## Example Workflow

1. Start the data-ingest service:

```bash
cd data-ingest
go run main.go
```

2. Start the trading-algo service:

```bash
cd trading-algo
go run main.go
```

3. View real-time data on WebSocket endpoints:

- Access real-time market data from the `data-ingest` service on `ws://localhost:8080/ws`.
- Access processed trading data (EMA/VWAP) from the `trading-algo` service on `ws://localhost:8090/ws`.

## Conclusion

This project demonstrates a basic setup where market data is ingested, processed with trading algorithms, and streamed via both gRPC and WebSocket servers. The EMA and VWAP calculations provide insights for executing a rubberband strategy.
