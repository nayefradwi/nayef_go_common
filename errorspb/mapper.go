package errorspb

import customerrors "github.com/nayefradwi/nayef_go_common/errors"

func FromResultError(e *customerrors.ResultError) *ResultErrorPb {
	if e == nil {
		return nil
	}
	pbErrors := make(map[string]*ErrorDetailsPbList, len(e.Errors))
	for field, details := range e.Errors {
		items := make([]*ErrorDetailsPb, len(details))
		for i, d := range details {
			items[i] = &ErrorDetailsPb{
				Message: d.Message,
				Code:    d.Code,
				Field:   d.Field,
			}
		}
		pbErrors[field] = &ErrorDetailsPbList{Items: items}
	}
	return &ResultErrorPb{
		Message: e.Message,
		Code:    e.Code,
		Errors:  pbErrors,
	}
}

// ToResultError converts a ResultErrorPb back to a ResultError.
func ToResultError(pb *ResultErrorPb) *customerrors.ResultError {
	if pb == nil {
		return nil
	}
	errs := make(map[string][]customerrors.ErrorDetails, len(pb.Errors))
	for field, list := range pb.Errors {
		if list == nil {
			continue
		}
		details := make([]customerrors.ErrorDetails, len(list.Items))
		for i, item := range list.Items {
			details[i] = customerrors.ErrorDetails{
				Message: item.Message,
				Code:    item.Code,
				Field:   item.Field,
			}
		}
		errs[field] = details
	}
	return &customerrors.ResultError{
		Message: pb.Message,
		Code:    pb.Code,
		Errors:  errs,
	}
}
