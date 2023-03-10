package nimona

//go:generate ./bin/mockgen -package=nimona -source=resolver.go -destination=resolver_mock.go

type Resolver interface {
	ResolveIdentityAlias(IdentityAlias) (*IdentityInfo, error)
}
