package common

import (
	"fmt"
	"os/exec"

	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func CheckBufInstalled() error {
	if _, err := exec.LookPath("buf"); err != nil {
		return fmt.Errorf("gRPC projects need the buf CLI on PATH, see https://buf.build/docs/cli/installation: %w", err)
	}

	return nil
}

func RunBufGenerate(dir string) error {
	stop := printer.Spin("Generating protobuf code")
	cmd := exec.Command("buf", "generate")
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	stop(err)
	if err != nil {
		return fmt.Errorf("failed to run buf generate %w: %s", err, out)
	}

	return nil
}
