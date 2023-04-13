package nimona

type (
	Profile struct {
		Metadata      Metadata            `nimona:"$metadata,omitempty,type=core/identity/profile"`
		KeygraphID    KeygraphID          `nimona:"keygraphID,omitempty"`
		IdentityAlias IdentityAlias       `nimona:"identityAlias,omitempty"`
		DisplayName   string              `nimona:"displayName,omitempty"`
		Repositories  []ProfileRepository `nimona:"repositories,omitempty"`
	}
	ProfileRepository struct {
		KeygraphID    KeygraphID `nimona:"keygraphID,omitempty"`
		Alias         string     `nimona:"alias,omitempty"`
		Handle        string     `nimona:"handle,omitempty"`
		DocumentTypes []string   `nimona:"documentTypes,omitempty"`
		// patch metadata
		Key       string   `nimona:"_key,omitempty"`
		Partition []string `nimona:"_partition,omitempty"`
	}
)
