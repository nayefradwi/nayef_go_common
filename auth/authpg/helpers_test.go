package authpg

import "github.com/google/uuid"

var testOwner = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var testTokenID = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")

func mustUUID(s string) uuid.UUID { return uuid.MustParse(s) }
