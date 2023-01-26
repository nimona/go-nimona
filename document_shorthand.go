package nimona

type Shorthand string

const (
	ShorthandDocumentID  Shorthand = "nimona://doc:"
	ShorthandPeerAddress Shorthand = "nimona://peer:addr:"
	ShorthandPeerKey     Shorthand = "nimona://peer:key:"

	ShorthandIdentity      Shorthand = "nimona://id:"       // ...<keystreamID>
	ShorthandIdentityAlias Shorthand = "nimona://id:alias:" // ...<handle>@<hostname>

	ShorthandNetwork      Shorthand = "nimona://net:"       // ...<keystreamID>
	ShorthandNetworkAlias Shorthand = "nimona://net:alias:" // ...<hostname>
)

func (t Shorthand) String() string {
	return string(t)
}
