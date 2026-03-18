package errorspb

func (err *ResultErrorPb) Error() string {
	return err.Message
}
