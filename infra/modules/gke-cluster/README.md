# GKE Cluster

Module Name: gke-cluster 
Language: Go
 
This module allows simplified creation and management of GKE clusters and Nodepools. Some sensible defaults are set initially, in order to allow less verbose usage for most use cases.


## Example

```
package main

import (
	gkecluster "github.com/google-cloud/google-cloud-pulumi/go/modules/gke-cluster"
    
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
			gke, err := gkecluster.NewContainerCluster(ctx, fmt.Sprintf("gke-%d", i), gkecluster.ContainerClusterArgs{
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
			fmt.Println(gke)
			fmt.Println(gke.Name)

			// Creat the NodePools
			_, err = gkecluster.NewGKENodePool(ctx, fmt.Sprintf("gke-np-%d", i), gkecluster.GKENodePoolArgs{
				ProjectId: project.ProjectId,
				Cluster:   gke.Name,
				Location:  gke.Location,
				NodeConfig: gkecluster.GKENodePoolNodeConfig{
					DiskSizeGb: 100,
				},
			}, nil)
			if err != nil {
				return err
			}
    })
}
