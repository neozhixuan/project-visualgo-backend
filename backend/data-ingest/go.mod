module github.com/neozhixuan/project-visualgo-backend/data-ingest

go 1.22.1

require (
	github.com/gorilla/websocket v1.5.1
	github.com/joho/godotenv v1.5.1
	github.com/neozhixuan/project-visualgo-backend/pb v0.0.0
	google.golang.org/grpc v1.63.2
)

require (
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/neozhixuan/project-visualgo-backend/pb => ../pb
