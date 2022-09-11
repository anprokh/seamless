package main

import (
	"fmt"
	"net/http"
	"os"
	"seamless/pkg/addendpoint"
	"seamless/pkg/addservice"
	"seamless/pkg/addtransport"
)

func main() {
	var (
		defaultUser     = os.Getenv("POSTGRES_USER")
		defaultPassword = os.Getenv("POSTGRES_PASSWORD")
		defaultDB       = os.Getenv("POSTGRES_DB")
		defaultHost     = os.Getenv("POSTGRES_HOST")
	)
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", defaultHost, defaultUser, defaultPassword, defaultDB)

	svc, err := addservice.NewSeamlessService(connStr)
	if err != nil {
		panic(err)
	}

	endpoints := addendpoint.MakeEndpoints(svc)
	jsonrpcHandler := addtransport.NewJSONRPCHandler(endpoints)

	http.Handle("/seamless", jsonrpcHandler)
	http.ListenAndServe(":8080", nil)
}
