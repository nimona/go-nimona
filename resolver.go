package nimona

//go:generate ./bin/mockgen -package=nimona -source=resolver.go -destination=resolver_mock.go

type Resolver interface {
	Resolve(NetworkID) ([]PeerAddr, error)
}
