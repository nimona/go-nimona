package keystream

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
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
//    Interaction event with the Inception object's hash and permissions in
//    its Seal.
// 9. The primary keystream controller sends the Delegation Interaction object
//    to the initiator peer for verification.
//
//    ╔════════════════════════╗                  ╔════════════════════════╗
//    ║       Initiator        ║                  ║       Delegator        ║
//    ╚═══════════╦════════════╝                  ╚════════════╦═══════════╝
//                │ ┌───────────────────┐                      │
//                │ │DelegationRequest  │                      │
//                │─┤-------------------├─────────────────────▶│
//                │ │* Permissions      │                      │
//                │ └───────────────────┘                      │
//                │               ┌──────────────────────────┐ │
//                │               │DelegationOffer           │ │
//                │◀──────────────┤--------------------------├─│
//      ┌─────────┴─────────┐     │* DelegatorSeal           │ │
//      │Inception          │     │  * DelegatorKeyStreamRoot│ │
//      │-------------------│     │  * Next stream sequence  │ │
//      │* DelegatorSeal    │     └──────────────────────────┘ │
//      └─────────┬─────────┘                                  │
//                │ ┌──────────────────────┐                   │
//                │ │DelegationVerification│                   │
//                │─┤----------------------├──────────────────▶│
//                │ │* DelegateSeal        │        ┌──────────┴──────────┐
//                │ └──────────────────────┘        │DelegationInteraction│
//                │                                 │---------------------│
//                │                                 │* DelegateSeal       │
//                │                                 └──────────┬──────────┘
//                │                                            │
//                ■                                            ■
//

type (
	DelegationRequest struct {
		Metadata                object.Metadata      `nimona:"@metadata:m,type=keystream.DelegationRequest"`
		InitiatorConnectionInfo *peer.ConnectionInfo `nimona:"initiatorConnectionInfo:m"`
		RequestedPermissions    Permissions          `nimona:"requestedPermissions:m"`
	}
	DelegationOffer struct {
		Metadata      object.Metadata `nimona:"@metadata:m,type=keystream.DelegationOffer"`
		DelegatorSeal DelegatorSeal   `nimona:"delegatorSeal:m"`
	}
	DelegationVerification struct {
		Metadata     object.Metadata `nimona:"@metadata:m,type=keystream.DelegationVerification"`
		DelegateSeal DelegateSeal    `nimona:"delegateSeal:m"`
	}
)

// NewDelegationRequest creates a new Delegation Request object.
func (m *manager) NewDelegationRequest(
	ctx context.Context,
	requestedPermissions Permissions,
) (*DelegationRequest, chan Controller, error) {
	ci := m.network.GetConnectionInfo()
	dr := &DelegationRequest{
		Metadata: object.Metadata{
			Owner: ci.PublicKey.DID(),
		},
		InitiatorConnectionInfo: ci,
		RequestedPermissions:    requestedPermissions,
	}

	res := make(chan Controller)

	go func() {
		defer close(res)
		// wait for a Delegation Offer
		env, err := m.network.SubscribeOnce(
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
		if dr.RequestedPermissions != do.DelegatorSeal.Permissions {
			return
		}
		// create new KeyStream
		ctrl, err := m.NewController(&do.DelegatorSeal)
		if err != nil {
			return
		}
		// send back the root hash of the keystream
		ver := &DelegationVerification{
			Metadata: object.Metadata{
				Owner: ci.PublicKey.DID(),
			},
			DelegateSeal: DelegateSeal{
				Root:        ctrl.GetKeyStream().Root,
				Permissions: requestedPermissions,
			},
		}
		verObj, err := object.Marshal(ver)
		if err != nil {
			return
		}
		err = m.network.Send(
			ctx,
			verObj,
			env.Sender,
		)
		if err != nil {
			return
		}
		// and pass the controller to the caller
		res <- ctrl
	}()

	return dr, res, nil
}

// HandleDelegationRequest handles a Delegation Request object.
//
// Given a Delegation Request object, creates a Delegation Offer, sends it to
// the Initiator peer, waits for the Inception event, and finally creates a
// Delegation Interaction event and stores it in the controller.
