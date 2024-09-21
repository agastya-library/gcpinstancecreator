package libs

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

// CreateVM creates a new VM instance based on the configuration
func CreateVM(ctx context.Context, creds *google.Credentials, config *Config) {
	client, err := compute.NewInstancesRESTClient(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}
	defer client.Close()

	machineTypePath := fmt.Sprintf("zones/%s/machineTypes/%s", config.Zone, config.MachineType)
	sourceImagePath := fmt.Sprintf("projects/%s/global/images/family/%s", config.ImageDetails.ImageProject, config.ImageDetails.ImageFamily)
	isBootDisk := true
	diskAutoDelete := true

	// Metadata for SSH keys
	metadataItems := []*computepb.Items{
		{
			Key:   proto.String("ssh-keys"),
			Value: proto.String(fmt.Sprintf("%s:%s", config.SSHDetails.Username, config.SSHDetails.PublicKey)),
		},
	}

	instance := &computepb.Instance{
		Name:        &config.InstanceName,
		MachineType: &machineTypePath,
		Disks: []*computepb.AttachedDisk{
			{
				AutoDelete: &diskAutoDelete,
				Boot:       &isBootDisk,
				InitializeParams: &computepb.AttachedDiskInitializeParams{
					SourceImage: &sourceImagePath,
				},
			},
		},
		Scheduling: &computepb.Scheduling{
			Preemptible: proto.Bool(true),
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				Name: new(string),
			},
		},
		Metadata: &computepb.Metadata{
			Items: metadataItems,
		},
	}

	req := &computepb.InsertInstanceRequest{
		Project:          config.ProjectID,
		Zone:             config.Zone,
		InstanceResource: instance,
	}

	op, err := client.Insert(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create instance: %v", err)
	}

	fmt.Printf("Instance creation in progress: %v\n", op)
}