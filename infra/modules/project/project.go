package project

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ProjectState struct {
	pulumi.ResourceState
}

type ProjectArgs struct {
	OrganizationId    pulumi.StringInput `pulumi:"organizationid"`
	FolderId          pulumi.StringInput `pulumi:"folderid"`
	Name              pulumi.StringInput `pulumi:"name"`
	ProjectId         pulumi.StringInput `pulumi:"projectid"`
	BillingAccount    pulumi.StringInput `pulumi:"billingaccount"`
	AutoCreateNetwork pulumi.Bool        `pulumi:"autocreatenetwork"`
	OSLogin           struct {
		Enabled pulumi.Bool        `pulumi:"enabled"`
		Admins  pulumi.StringInput `pulumi:"admins"`
		Users   pulumi.StringInput `pulumi:"users"`
	}
}

func NewProject(ctx *pulumi.Context, name string, args ProjectArgs, opts pulumi.ResourceOption) (*ProjectState, error) {
	project := &ProjectState{}
	err := ctx.RegisterComponentResource("pkg:google:project", name, project, opts)

	project, err := organizations.NewProject(ctx, name, &organizations.ProjectArgs{
		ProjectId:         args.ProjectId,
		BillingAccount:    args.BillingAccount,
		AutoCreateNetwork: args.AutoCreateNetwork,
	})
	if err != nil {
		return err
	}
}
