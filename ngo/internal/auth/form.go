package auth

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func RunForm() (*CreateAuthRequest, error) {
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

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[common.Identifier]().
				Title("Main identifier").
				Options(huh.NewOptions(common.IdentifierEmail, common.IdentifierPhone)...).
				Value(&req.Identifier),
			huh.NewSelect[common.LoginMethod]().
				Title("Login").
				Options(huh.NewOptions(common.LoginPassword, common.LoginOtp)...).
				Value(&req.Login),
			huh.NewConfirm().
				Title("Generate MFA (TOTP) endpoints?").
				Value(&req.Mfa),
			huh.NewSelect[common.AccessToken]().
				Title("Access token").
				Description("jwt: checked without a store read, lives until it expires. opaque: looked up per request, logout ends it at once").
				Options(huh.NewOptions(common.AccessTokenJwt, common.AccessTokenOpaque)...).
				Value(&req.AccessToken),
		),
		huh.NewGroup(
			huh.NewSelect[common.InfraType]().
				Title("Auth stores").
				Description("sessions, codes, attempts and TOTP; users always live in postgres").
				Options(huh.NewOptions(common.InfraTypePostgres, common.InfraTypeRedis)...).
				Value(&req.Store),
		).WithHideFunc(func() bool { return !req.HasRedis() }),
	)
	if err := form.Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return req, nil
}
