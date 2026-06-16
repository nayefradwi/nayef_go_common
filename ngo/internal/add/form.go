package add

import (
	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func RunForm() (*CreateFeatureRequest, error) {
	req := &CreateFeatureRequest{}

	form := huh.NewForm(
		huh.NewGroup(common.NameInput(&req.Name)),
		huh.NewGroup(common.InfraTypeInput(&req.InfraTypes)),
		huh.NewGroup(common.AuthInput(&req.AuthType)),
		huh.NewGroup(common.FeatureInput(&req.Features)),
	)
	if err := form.Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return req, nil
}
