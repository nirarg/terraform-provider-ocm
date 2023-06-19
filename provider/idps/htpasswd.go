package idps

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
)

type HTPasswdIdentityProvider struct {
	Users []HTPasswdUser `tfsdk:"users"`
}

type HTPasswdUser struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func HtpasswdSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"users": {
			Description: "List of users to add to the IDP.",
			Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
				"username": {
					Description: "User name.",
					Type:        types.StringType,
					Required:    true,
				},
				"password": {
					Description: "User password.",
					Type:        types.StringType,
					Required:    true,
					Sensitive:   true,
				},
			}, tfsdk.ListNestedAttributesOptions{},
			),
			Required: true,
		},
	})
}

func CreateHTPasswdIDPBuilder(ctx context.Context, state *HTPasswdIdentityProvider) *cmv1.HTPasswdUserListBuilder {

	users := state.Users

	htpasswdUsers := []*cmv1.HTPasswdUserBuilder{}

	if len(users) != 0 {
		//if a user list is specified then continue with the  list
		for _, user := range users {
			tflog.Info(ctx, fmt.Sprintf("*** Adding user %v", user))
			username := user.Username.Value
			password := user.Password.Value

			htpasswdUsers = append(htpasswdUsers, cmv1.NewHTPasswdUser().Username(username).Password(password))
		}
	}
	tflog.Info(ctx, fmt.Sprintf("Adding users %v", htpasswdUsers))

	htpassUserList := cmv1.NewHTPasswdUserList().Items(htpasswdUsers...)
	return htpassUserList
}

func UpdateHTPasswd(ctx context.Context, state *HTPasswdIdentityProvider, idpID string, idpClient *cmv1.IdentityProvidersClient) error {
	tflog.Info(ctx, fmt.Sprintf("******* state: %v", state))
	userList, err := CreateHTPasswdIDPBuilder(ctx, state).Build()
	if err != nil {
		return err
	}
	response, err := idpClient.IdentityProvider(idpID).HtpasswdUsers().Import().Items(userList.Slice()).Send()
	if err != nil {
		return common.HandleErr(response.Error(), err)
	}
	return nil
}
