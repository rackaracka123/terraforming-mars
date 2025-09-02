//go:generate swag init -g cmd/server/main.go -o ./docs/swagger
//go:generate go run tools/generate-types.go

package main