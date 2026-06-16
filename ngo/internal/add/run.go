package add

func Run() error {
	req, err := RunForm()

	if err != nil {
		return err
	}

	return generateServiceClass(*req)
}
