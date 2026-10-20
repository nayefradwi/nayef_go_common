package auth

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
	"github.com/spf13/pflag"
)

func RunForm(flags *pflag.FlagSet) (*CreateAuthRequest, error) {
	wd, err := os.Getwd()
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	project, err := common.ReadProject(wd)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	goModule, err := common.ReadGoModule(wd)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	req := &CreateAuthRequest{
		Identifier:  common.IdentifierEmail,
		Login:       common.LoginPassword,
		AccessToken: common.AccessTokenJwt,
		Store:       common.InfraTypePostgres,
		InfraTypes:  project.InfraTypes,
		ServiceType: project.ServiceType,
		GoModule:    goModule,
		RootDirPath: wd,
		Project:     project,
	}

	groups, err := fillFromFlags(req, flags)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return req, nil
}

func fillFromFlags(req *CreateAuthRequest, flags *pflag.FlagSet) ([]*huh.Group, error) {
	stores := []common.InfraType{common.InfraTypePostgres}
	if req.HasRedis() {
		stores = append(stores, common.InfraTypeRedis)
	}

	return common.FillFromFlags(flags, []common.FlagField{
		{
			Name: FLAG_IDENTIFIER,
			Set: func() (err error) {
				req.Identifier, err = common.ParseFlag(flags, FLAG_IDENTIFIER, AllIdentifiers)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(
					huh.NewSelect[common.Identifier]().
						Title("Main identifier").
						Options(huh.NewOptions(AllIdentifiers...)...).
						Value(&req.Identifier),
				)
			},
		},
		{
			Name: FLAG_LOGIN,
			Set: func() (err error) {
				req.Login, err = common.ParseFlag(flags, FLAG_LOGIN, AllLoginMethods)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(
					huh.NewSelect[common.LoginMethod]().
						Title("Login").
						Options(huh.NewOptions(AllLoginMethods...)...).
						Value(&req.Login),
				)
			},
		},
		{
			Name: FLAG_MFA,
			Set: func() (err error) {
				req.Mfa, err = flags.GetBool(FLAG_MFA)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(
					huh.NewConfirm().
						Title("Generate MFA (TOTP) endpoints?").
						Value(&req.Mfa),
				)
			},
		},
		{
			Name: FLAG_ACCESS_TOKEN,
			Set: func() (err error) {
				req.AccessToken, err = common.ParseFlag(flags, FLAG_ACCESS_TOKEN, AllAccessTokens)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(
					huh.NewSelect[common.AccessToken]().
						Title("Access token").
						Description("jwt: checked without a store read, lives until it expires. opaque: looked up per request, logout ends it at once").
						Options(huh.NewOptions(AllAccessTokens...)...).
						Value(&req.AccessToken),
				)
			},
		},
		{
			Name: FLAG_STORE,
			Set: func() (err error) {
				req.Store, err = common.ParseFlag(flags, FLAG_STORE, stores)
				return
			},
			Group: func() *huh.Group {
				if len(stores) == 1 {
					return nil
				}
				return huh.NewGroup(
					huh.NewSelect[common.InfraType]().
						Title("Auth stores").
						Description("sessions, codes, attempts and TOTP; users always live in postgres").
						Options(huh.NewOptions(stores...)...).
						Value(&req.Store),
				)
			},
		},
	})
}
