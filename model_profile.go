package nimona

type (
	Profile struct {
		Metadata      Metadata            `nimona:"$metadata,omitempty,type=core/identity/profile"`
		Identity      Identity            `nimona:"identity,omitempty"`
		IdentityAlias *IdentityAlias      `nimona:"identityAlias,omitempty"`
		DisplayName   string              `nimona:"displayName,omitempty"`
		Repositories  []ProfileRepository `nimona:"repositories,omitempty"`
	}
	ProfileRepository struct {
		Identity      Identity `nimona:"identity,omitempty"`
		Alias         string   `nimona:"alias,omitempty"`
		Handle        string   `nimona:"handle,omitempty"`
		DocumentTypes []string `nimona:"documentTypes"`
	}
)
