package new

import (
	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func RunForm() error {
	name := new("")
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Project Name").Placeholder("my-service").Value(name),
		),
	)

	if err := form.Run(); err != nil {
		printer.Error(err.Error())
		return err
	}

	return nil
}
