package validation

type testUser struct {
	name  string
	email *string
	age   int
}

func (u testUser) validate() error {
	validator := NewValidator()
	strValidator := StringValidatorFromValidator(validator)
	strValidator.IsRequired(ValidateOptionsFrom(u.name, "name", "Name is required"))
	strValidator.IsEmail(ValidateOptionsFrom(u.email, "email", "Email is required"))

	return validator.Validate()
}
