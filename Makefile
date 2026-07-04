test:
	go test ./...

build:
	go build -o bin/node cmd/node/main.go

up:
	go run cmd/node/main.go

load:
	go run cmd/wallet/main.go send --from 0x32Df1b36e74cCf1c3c987B151650E1F1170B5258 --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 1
	go run cmd/wallet/main.go send --from 0x32Df1b36e74cCf1c3c987B151650E1F1170B5258 --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 2
	go run cmd/wallet/main.go send --from 0x32Df1b36e74cCf1c3c987B151650E1F1170B5258 --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 3
	go run cmd/wallet/main.go send --from 0x32Df1b36e74cCf1c3c987B151650E1F1170B5258 --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 4
	go run cmd/wallet/main.go send --from 0x32Df1b36e74cCf1c3c987B151650E1F1170B5258 --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 5
	go run cmd/wallet/main.go send --from 0x32Df1b36e74cCf1c3c987B151650E1F1170B5258 --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 6
