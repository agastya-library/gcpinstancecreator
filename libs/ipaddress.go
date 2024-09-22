package libs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
//	"github.com/davecgh/go-spew/spew"
	"github.com/googleapis/gax-go/v2/apierror"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
//	"google.golang.org/grpc/codes"
//	"google.golang.org/grpc/status"
)

// ReserveNewGlobalExternal reserves a new global external IP address or retrieves an existing one.
func ReserveNewGlobalExternal(w io.Writer, projectID string, ipDets *IpDetails, creds *google.Credentials) (*computepb.Address, error) {
	ctx := context.Background()

	client, err := compute.NewGlobalAddressesRESTClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("NewGlobalAddressesRESTClient: %w", err)
	}
	defer client.Close()

	// Check if the address already exists
	existingAddress, err := client.Get(ctx, &computepb.GetGlobalAddressRequest{
		Project: projectID,
		Address: ipDets.Name,
	})
	if err == nil {
		fmt.Fprintf(w, "Global address %v already exists: %v\n", ipDets.Name, existingAddress.GetAddress())
		return existingAddress, nil
	} else if !isNotFoundError(err) {
		return nil, fmt.Errorf("error checking for existing global address: %w", err)
	}


	ipv6EndPointType := "VM"
	purpose := "NAT_AUTO"
	address := &computepb.Address{
		Name:      &ipDets.Name,
		IpVersion: &ipDets.IpVersion,
		Ipv6EndpointType: &ipv6EndPointType,
		Region: &ipDets.Region,
		NetworkTier:  &ipDets.NetworkTier,
		Purpose: &purpose,
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
		Address: ipDets.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get reserved global address: %w", err)
	}

	fmt.Fprintf(w, "New global address %v reserved: %v\n", ipDets.Name, newAddress.GetAddress())
	return newAddress, nil
}

// Helper function to check if the error is a "not found" error
func isNotFoundError(err error) bool {
	// Convert the error to a gRPC status error and check if it is codes.NotFound
    if apiErr, ok := err.(*apierror.APIError); ok {
        // Now you can use apiErr as an apierror.APIError
		return apiErr.HTTPCode() == 404
    } else {
        fmt.Printf("Other error: %s\n", err)
		return false
    }
}