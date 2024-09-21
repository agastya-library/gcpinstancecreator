package main

import (
	"context"
	"fmt"
	"log"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/agastya-library/gcpinstancecreator/libs"
	"golang.org/x/oauth2/google"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config-file>", os.Args[0])
	}

	// Load configuration
	configFile := os.Args[1]
	conf, err :=  libs.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Load credentials
	credentialsData, err := os.ReadFile(conf.Credentials)
	if err != nil {
		log.Fatalf("Failed to read credentials file: %v", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, credentialsData, compute.DefaultAuthScopes()...)
	if err != nil {
		log.Fatalf("Failed to get credentials from JSON: %v", err)
	}

	// Reserve external IP
	addressName := conf.IpDetails.Name // Change this as needed
	isV6 := conf.IpDetails.IpV6
	address, err :=  libs.ReserveNewGlobalExternal(os.Stdout, conf.ProjectID, addressName, isV6, creds)
	if err != nil {
		fmt.Printf("Error reserving global external address: %v\n", err)
		return
	}

	fmt.Printf("Successfully reserved or retrieved global address: %v\n", address.GetAddress())

	// Create VM instance
	libs.CreateVM(ctx,creds,conf)
}