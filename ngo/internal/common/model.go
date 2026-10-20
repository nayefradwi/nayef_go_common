package common

import (
	"slices"
)

type ServiceType string

const (
	ServiceTypeRest ServiceType = "REST"
	ServiceTypeGrpc ServiceType = "gRPC"
)

var AllServiceTypes = []ServiceType{ServiceTypeRest, ServiceTypeGrpc}

type InfraType string

const (
	InfraTypePostgres InfraType = "postgresql"
	InfraTypeRedis    InfraType = "redis"
)

var AllInfraTypes = []InfraType{InfraTypePostgres, InfraTypeRedis}

type Feature string

const (
	FeatureLocking    Feature = "Locking"
	FeaturePagination Feature = "Pagination"
)

var AllFeatures = []Feature{FeatureLocking, FeaturePagination}

type Identifier string

const (
	IdentifierEmail Identifier = "email"
	IdentifierPhone Identifier = "phone"
)

type LoginMethod string

const (
	LoginPassword LoginMethod = "password"
	LoginOtp      LoginMethod = "otp"
)

type AccessToken string

const (
	AccessTokenJwt    AccessToken = "jwt"
	AccessTokenOpaque AccessToken = "opaque"
)

type Auth struct {
	Identifier  Identifier  `yaml:"identifier"`
	Login       LoginMethod `yaml:"login"`
	Mfa         bool        `yaml:"mfa"`
	AccessToken AccessToken `yaml:"access_token"`
	Store       InfraType   `yaml:"store"`
}

type TakesFeatures struct {
	Features []Feature
}

type TakesInfraTypes struct {
	InfraTypes []InfraType
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
