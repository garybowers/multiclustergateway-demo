// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gkecluster

import (
	"fmt"
	"infra/utils/defaulter"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type GKEState struct {
	pulumi.ResourceState
	Name     pulumi.StringOutput `pulumi:"name"`
	Location pulumi.StringOutput `pulumi:"location"`
}

type GKEArgs struct {
	ProjectId   pulumi.StringInput `pulumi:"projectId"`
	Location    pulumi.StringInput `pulumi:"location"`
	Name        pulumi.StringInput `pulumi:"name"`
	Description pulumi.StringInput `pulumi:"description"`
	AutoPilot   pulumi.Bool
	NetConfig   GKENetworkConfig
}
type GKENetworkConfig struct {
	Network    pulumi.StringInput `pulumi:"network"`
	SubNetwork pulumi.StringInput `pulumi:"subnetwork"`
}

func NewGKECluster(ctx *pulumi.Context, name string, args GKEArgs, opts pulumi.ResourceOption) (*GKEState, error) {
	gkeCluster := &GKEState{}
	err := ctx.RegisterComponentResource("pkg:google:gke-cluster", name, gkeCluster, opts)
	if err != nil {
		return nil, err
	}

	gke, err := container.NewCluster(ctx, name, &container.ClusterArgs{
		Project:               args.ProjectId,
		Location:              args.Location,
		RemoveDefaultNodePool: pulumi.Bool(true),
		InitialNodeCount:      pulumi.Int(1),
		Network:               args.NetConfig.Network,
		Subnetwork:            args.NetConfig.SubNetwork,
	})
	if err != nil {
		return nil, err
	}

	gkeCluster.Name = gke.Name
	gkeCluster.Location = gke.Location
	return gkeCluster, nil
}

/*
================================
	GKE Nodepool
================================
*/

type GKENodePoolState struct {
	pulumi.ResourceState
}

type GKENodePoolArgs struct {
	ProjectId  pulumi.StringInput    `pulumi:"projectId"`
	Location   pulumi.StringInput    `pulumi:"location"`
	Cluster    pulumi.StringInput    `pulumi:"cluster"`
	Name       pulumi.StringInput    `pulumi:"name"`
	NodeConfig GKENodePoolNodeConfig `pulumi:"nodeconfig"`
	/*
		NodeConfig struct {
			MachineType pulumi.StringInput `pulumi:"machinetype" default:"e2-standard"`
			DiskSizeGb  pulumi.Int         `pulumi:"disksizegb" default:100`
			DiskType    pulumi.StringInput `pulumi:"disktype" default:"PD-BALANCED"`
		}
	*/
}

type GKENodePoolNodeConfig struct {
	MachineType pulumi.StringInput `pulumi:"machinetype" default:"e2-standard"`
	DiskSizeGb  pulumi.Int         `pulumi:"disksizegb" default:"90"`
	DiskType    pulumi.StringInput `pulumi:"disktype" default:"PD-STANDARD"`
}

func NewGKENodePool(ctx *pulumi.Context, name string, args GKENodePoolArgs, opts pulumi.ResourceOption) (*GKENodePoolState, error) {
	gkeNodePool := &GKENodePoolState{}

	err := defaulter.SetDefaults(&args)
	if err != nil {
		return nil, err
	}

	err = ctx.RegisterComponentResource("pkg:google:gke-nodepool", name, gkeNodePool, opts)
	if err != nil {
		return nil, err
	}

	_, err = serviceaccount.NewAccount(ctx, name, &serviceaccount.AccountArgs{
		Project:     args.ProjectId,
		AccountId:   pulumi.String(fmt.Sprintf("svc-%v", name)),
		DisplayName: pulumi.String(name),
	})
	if err != nil {
		return nil, err
	}

	_, err = container.NewNodePool(ctx, name, &container.NodePoolArgs{
		Project:  args.ProjectId,
		Location: args.Location,
		Cluster:  args.Cluster,
		Name:     args.Name,
		NodeConfig: &container.NodePoolNodeConfigArgs{
			MachineType: args.NodeConfig.MachineType,
			DiskSizeGb:  args.NodeConfig.DiskSizeGb, // cannot be 0
			DiskType:    args.NodeConfig.DiskType,
		},
	})
	if err != nil {
		return nil, err
	}
	return gkeNodePool, nil
}
