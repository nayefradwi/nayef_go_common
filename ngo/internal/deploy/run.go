package deploy

func Run() error {
	req, err := RunForm()
	if err != nil {
		return err
	}

	return generateDeployment(*req)
}
