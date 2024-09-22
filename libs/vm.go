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
func CreateVM(ctx context.Context, creds *google.Credentials, config *Config, reservedIp string) {
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
					DiskSizeGb: &config.DiskSize,
				},
			},
		},
		Scheduling: &computepb.Scheduling{
			Preemptible: proto.Bool(true),
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
                AccessConfigs: []*computepb.AccessConfig{
                    {
                        Name:  proto.String("External IPv6"),
						NetworkTier: &config.NetworkTier,
                    },
                },
				Ipv6AccessConfigs: []*computepb.AccessConfig{
					{
						Name: proto.String("External IP V6"),
						ExternalIpv6: &reservedIp,
						ExternalIpv6PrefixLength: proto.Int32(96),
						NetworkTier: &config.NetworkTier,
						Type: proto.String("DIRECT_IPV6"),
					},
				},
				StackType: proto.String("IPV4_IPV6"),
				Subnetwork: proto.String(fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s",config.ProjectID,config.NetworkRegion,config.NetworkSubnet)),
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