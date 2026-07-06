test:
	go test ./...

build:
	go build -o bin/node cmd/node/main.go

up:
	BENEFICIARY=miner1 HOST=0.0.0.0:8080 DB_PATH=zblock/miner1 go run cmd/node/main.go

up2:
	BENEFICIARY=miner2 HOST=0.0.0.0:8081 DB_PATH=zblock/miner2 go run cmd/node/main.go

load:
	go run cmd/wallet/main.go send --from 0xc1b0e54BAAc9C7eff1ca8F584534E13DA27ca5cc --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 1
	go run cmd/wallet/main.go send --from 0xc1b0e54BAAc9C7eff1ca8F584534E13DA27ca5cc --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 2
	go run cmd/wallet/main.go send --from 0xc1b0e54BAAc9C7eff1ca8F584534E13DA27ca5cc --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 3

load2:
	go run cmd/wallet/main.go send --from 0xc1b0e54BAAc9C7eff1ca8F584534E13DA27ca5cc --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 4
	go run cmd/wallet/main.go send --from 0xc1b0e54BAAc9C7eff1ca8F584534E13DA27ca5cc --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 5

load3:
	go run cmd/wallet/main.go send --from 0xc1b0e54BAAc9C7eff1ca8F584534E13DA27ca5cc --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 6
	go run cmd/wallet/main.go send --from 0xc1b0e54BAAc9C7eff1ca8F584534E13DA27ca5cc --to 0x67e19dff8D01DE05038d5c7B8fbAF23dE9d302a7 --value 100 --tip 10 --nonce 7
