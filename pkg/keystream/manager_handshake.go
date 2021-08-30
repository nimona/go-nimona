package keystream

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

// Handshake allows a peer to initiate a request for a new delegated key stream
// from a remote peer holding the delegating key stream controller.
//
// The process of a handshake is as follows:
// 1. Initiator creates a new Delegation Request object that contains the peer's
//    public key and the requested permissions and signs it.
// 2. The object is encoded into a QR code or other suitable medium and should
//    be presented to the user.
// 3. The user scans the QR code with an identity application which has the
//    user's identity primary key stream, verifies the requested permissions
//    and confirms the request.
// 4. The primary keystream controller creates a Delegation Offer object that
//    contains the following:
//    - The stream root hash of the identity keystream
//    - The sequence of the next available object in the stream which will be
//      used for the DelegationInteraction object
//    - The permissions granted by the user
// 5. The Delegation Offer is signed, and sent to the initiator peer.
// 6. The initiator peer receives the Delegation Offer and creates and signs
//    an Inception object, that includes a Delegator Seal with the information
//    received in the Offer.
// 7. The initiator peer sends the Inception object to the primary keystream
//    controller.
// 8. The primary keystream controller receives the Inception object, checks
//    that the Delegator Seal matches the one sent, and creates a Delegation
//    Interaction object that contains the Inception object with the Inception
//    object's hash and permissions in the Seal.
// 9. The primary keystream controller sends the Delegation Interaction object
//    to the initiator peer for verification.

type (
	DelegationRequest struct {
		Metadata                object.Metadata      `nimona:"@metadata:m,type=keystream.DelegationRequest"`
		InitiatorConnectionInfo *peer.ConnectionInfo `nimona:"initiatorConnectionInfo:m"`
		RequestedPermissions    Permissions          `nimona:"requestedPermissions:m"`
	}
	DelegationOffer struct {
		Metadata                      object.Metadata `nimona:"@metadata:m,type=keystream.DelegationOffer"`
		DelegatorKeyStreamRoot        tilde.Digest    `nimona:"delegatorKeyStreamRoot:m"`
		DelegationInteractionSequence uint64          `nimona:"delegationInteractionSequence:i"`
		GrantedPermissions            Permissions     `nimona:"grantedPermissions:m"`
	}
)

// InitiateDelegationRequest creates a new Delegation Request object.
func InitiateDelegationRequest(
	ctx context.Context,
	net network.Network,
	requestedPermissions Permissions,
) (*DelegationRequest, error) {
	ci := net.GetConnectionInfo()
	dr := &DelegationRequest{
		Metadata: object.Metadata{
			Owner: ci.PublicKey.DID(),
		},
		InitiatorConnectionInfo: ci,
		RequestedPermissions:    requestedPermissions,
	}

	ks := make(chan Controller)

	go func() {
		defer close(ks)
		// wait for a Delegation Offer
		env, err := net.SubscribeOnce(
			ctx,
			network.FilterByObjectType("keystream.DelegationOffer"),
		)
		if err != nil {
			return
		}
		do := &DelegationOffer{}
		err = object.Unmarshal(env.Payload, do)
		if err != nil {
			return
		}
		// verify Offer
		if dr.RequestedPermissions != do.GrantedPermissions {
			return
		}
		// create new KeyStream
		// ctrl, err := NewController()
	}()

	return dr, nil
}

// HandleDelegationRequest handles a Delegation Request object.
//
// Given a Delegation Request object, creates a Delegation Offer, sends it to
// the Initiator peer, waits for the Inception event, and finally creates a
// Delegation Interaction event and stores it in the controller.
