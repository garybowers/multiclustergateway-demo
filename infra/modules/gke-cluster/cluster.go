package gkecluster

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ContainerClusterState struct {
	pulumi.ResourceState
}

type ContainerClusterArgs struct {
	ProjectId pulumi.StringInput `pulumi:"projectId"`
	Region    pulumi.StringInput `pulumi:"region"`
	VpcConfig ContainerClusterVpcConfig
}

type ContainerClusterVpcConfig struct {
	Network    pulumi.StringInput `pulumi:"network"`
	SubNetwork pulumi.StringInput `pulumi:"subnetwork"`
}

func NewContainerCluster(ctx *pulumi.Context, name string, args ContainerClusterArgs, opts pulumi.ResourceOption) (*ContainerClusterState, error) {
	containerCluster := &ContainerClusterState{}
	err := ctx.RegisterComponentResource("pkg:google:ContainerCluster", name, containerCluster, opts)
	if err != nil {
		return nil, err
	}

	svcAcc, err := serviceaccount.NewAccount(ctx, name, &serviceaccount.AccountArgs{
		Project:     args.ProjectId,
		AccountId:   pulumi.String(name),
		DisplayName: pulumi.String(name),
	})
	if err != nil {
		return nil, err
	}

	primary, err := container.NewCluster(ctx, name, &container.ClusterArgs{
		Project:               args.ProjectId,
		Location:              args.Region,
		RemoveDefaultNodePool: pulumi.Bool(true),
		InitialNodeCount:      pulumi.Int(1),
		Network:               args.VpcConfig.Network,
		Subnetwork:            args.VpcConfig.SubNetwork,
	})
	if err != nil {
		return nil, err
	}
	_, err = container.NewNodePool(ctx, name, &container.NodePoolArgs{
		Project:   args.ProjectId,
		Location:  pulumi.String("us-central1"),
		Cluster:   primary.Name,
		NodeCount: pulumi.Int(1),
		NodeConfig: &container.NodePoolNodeConfigArgs{
			Preemptible:    pulumi.Bool(true),
			MachineType:    pulumi.String("e2-medium"),
			ServiceAccount: svcAcc.Email,
			OauthScopes: pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/cloud-platform"),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return containerCluster, nil
}
