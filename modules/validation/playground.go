package validation

type testUser struct {
	name  string
	email string
	age   int
}

func (u testUser) validate() error {
	validator := NewValidator()

	validator.AddValidation(NotEmptyString(ValidateOptionsFrom(u.name, "name", "Name is required")))

	return validator.Validate()
}
