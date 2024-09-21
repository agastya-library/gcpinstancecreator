package libs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"golang.org/x/oauth2/google"
)

// ReserveNewGlobalExternal reserves a new global external IP address or retrieves an existing one.
func ReserveNewGlobalExternal(w io.Writer, projectID, addressName string, isV6 bool, creds *google.Credentials) (*computepb.Address, error) {
	ctx := context.Background()

	client, err := compute.NewGlobalAddressesRESTClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("NewGlobalAddressesRESTClient: %w", err)
	}
	defer client.Close()

	// Check if the address already exists
	existingAddress, err := client.Get(ctx, &computepb.GetGlobalAddressRequest{
		Project: projectID,
		Address: addressName,
	})
	if err == nil {
		fmt.Fprintf(w, "Global address %v already exists: %v\n", addressName, existingAddress.GetAddress())
		return existingAddress, nil
	} else if !isNotFoundError(err) {
		return nil, fmt.Errorf("error checking for existing global address: %w", err)
	}

	// If address doesn't exist, reserve a new one
	ipVersion := computepb.Address_IPV4.String()
	if isV6 {
		ipVersion = computepb.Address_IPV6.String()
	}

	address := &computepb.Address{
		Name:      &addressName,
		IpVersion: &ipVersion,
	}

	req := &computepb.InsertGlobalAddressRequest{
		Project:         projectID,
		AddressResource: address,
	}

	op, err := client.Insert(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to reserve global address: %w", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("waiting for global address reservation operation to complete: %w", err)
	}

	newAddress, err := client.Get(ctx, &computepb.GetGlobalAddressRequest{
		Project: projectID,
		Address: addressName,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get reserved global address: %w", err)
	}

	fmt.Fprintf(w, "New global address %v reserved: %v\n", addressName, newAddress.GetAddress())
	return newAddress, nil
}

// Helper function to check if the error is a "not found" error
func isNotFoundError(err error) bool {
	if apiErr, ok := err.(*googleapi.Error); ok {
		return apiErr.Code == 404
	}
	return false
}