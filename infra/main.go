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

package main

import (
	"fmt"
	gkecluster "infra/modules/gke-cluster"
	cloudnat "infra/modules/net-cloudnat"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// Get Stack Configs
		conf := config.New(ctx, "")

		orgId := conf.Require("orgId")
		folderName := conf.Require("folderName")
		billingAccount := conf.Require("billingAC")
		billingSvcAcc := conf.Require("billingSA")

		// Get a access token to assume identity to a service account that has the ability
		// To provision & attach projects to a billing account.
		//
		//   - Prerequisites:  a service account in a project with projectCreator and
		//                     billingAccountUser permissions on the organization

		accessToken, err := serviceaccount.GetAccountAccessToken(ctx, &serviceaccount.GetAccountAccessTokenArgs{
			TargetServiceAccount: billingSvcAcc,
			Scopes:               []string{"cloud-platform"},
		})
		if err != nil {
			return err
		}

		// Create provider config for billing account user
		googleBillingUser, err := gcp.NewProvider(ctx, "googlebillinguser", &gcp.ProviderArgs{
			AccessToken: pulumi.String(accessToken.AccessToken),
		})

		// Create a folder for the projects to reside in our org.
		folder, err := organizations.NewFolder(ctx, folderName, &organizations.FolderArgs{
			DisplayName: pulumi.String(folderName),
			Parent:      pulumi.String("organizations/" + orgId),
		})
		if err != nil {
			return err
		}

		var project *organizations.Project
		// Create the project using our credentials from the above provider configuration
		project, err = organizations.NewProject(ctx, "mcg-demo", &organizations.ProjectArgs{
			ProjectId:         pulumi.String("mcg-demo-h38hr1"),
			FolderId:          folder.Name,
			AutoCreateNetwork: pulumi.Bool(false),
			BillingAccount:    pulumi.String(billingAccount),
		}, pulumi.Provider(googleBillingUser))
		if err != nil {
			return err
		}

		// Enable services on the account
		services := []string{
			"container.googleapis.com",
			"artifactregistry.googleapis.com",
			"dns.googleapis.com"}

		for i, service := range services {
			_, err = projects.NewService(ctx, fmt.Sprintf("api-%d", i), &projects.ServiceArgs{
				DisableDependentServices: pulumi.Bool(false),
				Project:                  project.ProjectId,
				Service:                  pulumi.String(service),
			})
			if err != nil {
				return err
			}
		}

		// Create the VPC
		vpc, err := compute.NewNetwork(ctx, "vpc", &compute.NetworkArgs{
			AutoCreateSubnetworks: pulumi.Bool(false),
			Project:               project.ProjectId,
			RoutingMode:           pulumi.String("GLOBAL"),
		})
		if err != nil {
			return err
		}

		regions := []string{
			"europe-west1",
		}

		for i, region := range regions {
			// Create the required subnetworks
			sn, err := compute.NewSubnetwork(ctx, fmt.Sprintf("sn-%d", i), &compute.SubnetworkArgs{
				Project:               project.ProjectId,
				Network:               vpc.SelfLink,
				Region:                pulumi.String(region),
				IpCidrRange:           pulumi.String(fmt.Sprintf("10.%d.0.0/16", i)),
				PrivateIpGoogleAccess: pulumi.Bool(true),
			})
			if err != nil {
				return err
			}

			// Create the GKE Clusters in each region
			gke, err := gkecluster.NewGKECluster(ctx, fmt.Sprintf("gke-%d", i), gkecluster.GKEArgs{
				ProjectId: project.ProjectId,
				Location:  pulumi.String(region),
				NetConfig: gkecluster.GKENetworkConfig{
					Network:    vpc.SelfLink,
					SubNetwork: sn.SelfLink,
				},
			}, nil)
			if err != nil {
				return err
			}

			// Creat the NodePools
			_, err = gkecluster.NewGKENodePool(ctx, fmt.Sprintf("gke-np-%d", i), gkecluster.GKENodePoolArgs{
				ProjectId: project.ProjectId,
				Cluster:   gke.Name,
				Location:  gke.Location,

				NodeConfig: gkecluster.GKENodePoolNodeConfig{
					DiskSizeGb: 80,
				},
			}, nil)
			if err != nil {
				return err
			}

			// Create a NAT Router in each region
			cloudnat, err := cloudnat.NewNetCloudNat(ctx, fmt.Sprintf("cn-%d-%v", i, region), cloudnat.NetCloudNatArgs{
				ProjectId:  project.ProjectId,
				Region:     pulumi.String(region),
				VpcNetwork: vpc.SelfLink,
			}, nil)
			if err != nil {
				return err
			}
			fmt.Println(cloudnat)
		}

		return nil
	})
}
