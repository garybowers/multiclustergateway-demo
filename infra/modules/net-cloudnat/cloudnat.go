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

package netcloudnat

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type NetCloudNatState struct {
	pulumi.ResourceState
}

type NetCloudNatArgs struct {
	ProjectId  pulumi.StringInput `pulumi:"projectId"`
	Region     pulumi.StringInput `pulumi:"region"`
	VpcNetwork pulumi.StringInput `pulumi:"vpcnetwork"`
}

func NewNetCloudNat(ctx *pulumi.Context, name string, args NetCloudNatArgs, opts pulumi.ResourceOption) (*NetCloudNatState, error) {
	cloudnat := &NetCloudNatState{}

	err := ctx.RegisterComponentResource("pkg:google:NetCloudNat", name, cloudnat, opts)
	if err != nil {
		return nil, err
	}

	router, err := compute.NewRouter(ctx, name, &compute.RouterArgs{
		Project: args.ProjectId,
		Region:  args.Region,
		Network: args.VpcNetwork,
		Bgp: &compute.RouterBgpArgs{
			Asn: pulumi.Int(64514),
		},
	})
	if err != nil {
		return nil, err
	}
	_, err = compute.NewRouterNat(ctx, name, &compute.RouterNatArgs{
		Project:                       args.ProjectId,
		Region:                        args.Region,
		Router:                        router.Name,
		NatIpAllocateOption:           pulumi.String("AUTO_ONLY"),
		SourceSubnetworkIpRangesToNat: pulumi.String("ALL_SUBNETWORKS_ALL_IP_RANGES"),
		LogConfig: &compute.RouterNatLogConfigArgs{
			Enable: pulumi.Bool(true),
			Filter: pulumi.String("ERRORS_ONLY"),
		},
	})
	if err != nil {
		return nil, err
	}

	return cloudnat, nil
}
