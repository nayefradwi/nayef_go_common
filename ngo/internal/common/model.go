package common

import (
	"slices"
)

type InfraType string

const (
	InfraTypePostgres InfraType = "postgresql"
	InfraTypeRedis    InfraType = "redis"
)

type Feature string

const (
	FeatureLocking    Feature = "Locking"
	FeatureOtp        Feature = "OTP"
	FeaturePagination Feature = "Pagination"
)

type AuthType string

const (
	AuthTypeJWT     AuthType = "JWT"
	AuthTypeRefresh AuthType = "JWT + Revokable Refresh"
	AuthTypeOpaque  AuthType = "Opaque Tokens"
	AuthTypeNone    AuthType = "Something Else"
)

type TakesFeatures struct {
	Features []Feature
}

type TakesInfraTypes struct {
	InfraTypes []InfraType
}

type TakesAuthType struct {
	AuthType AuthType
}

func (r TakesInfraTypes) HasPostgres() bool {
	return slices.Contains(r.InfraTypes, InfraTypePostgres)
}

func (r TakesInfraTypes) HasRedis() bool {
	return slices.Contains(r.InfraTypes, InfraTypeRedis)
}

func (r TakesFeatures) HasFeature(f Feature) bool {
	return slices.Contains(r.Features, f)
}

func (r TakesAuthType) HasAuth() bool {
	return r.AuthType != AuthTypeNone
}
