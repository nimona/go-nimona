package nimona

type Shorthand string

// TODO(@geoah): consider removing addr/key/alias variants as discussed with @jimeh
const (
	ShorthandDocumentID Shorthand = "nimona://doc:" // ...<documentID>

	ShorthandPeerAddress Shorthand = "nimona://peer:addr:" // ...<transport>:<address>
	ShorthandPeerKey     Shorthand = "nimona://peer:key:"  // ...<publicKey>

	ShorthandIdentity      Shorthand = "nimona://id:"       // ...<keystreamID>
	ShorthandIdentityAlias Shorthand = "nimona://id:alias:" // ...<handle>@<hostname>

	ShorthandNetwork      Shorthand = "nimona://net:"       // ...<keystreamID>
	ShorthandNetworkAlias Shorthand = "nimona://net:alias:" // ...<hostname>
)

func (t Shorthand) String() string {
	return string(t)
}
