module nimona.io

go 1.19

require (
	github.com/fxamacker/cbor/v2 v2.4.0
	github.com/golang/mock v1.7.0-rc.1
	github.com/hashicorp/golang-lru/v2 v2.0.1
	github.com/ipfs/go-cid v0.3.2
	github.com/mr-tron/base58 v1.2.0
	github.com/neilalexander/utp v0.1.0
	github.com/oasisprotocol/curve25519-voi v0.0.0-20221003100820-41fad3beba17
	github.com/stretchr/testify v1.8.1
	github.com/whyrusleeping/cbor-gen v0.0.0-20221220214510-0333c149dec0
	golang.org/x/crypto v0.4.0
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
)

require (
	github.com/anacrolix/envpprof v1.2.1 // indirect
	github.com/anacrolix/missinggo v1.3.0 // indirect
	github.com/anacrolix/missinggo/perf v1.0.0 // indirect
	github.com/anacrolix/missinggo/v2 v2.5.1 // indirect
	github.com/anacrolix/sync v0.4.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.2 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.1.1 // indirect
	github.com/multiformats/go-multihash v0.2.1 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/sys v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
)

replace github.com/whyrusleeping/cbor-gen => ../github.com/geoah/go-cborgen
