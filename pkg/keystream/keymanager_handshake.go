package keystream

import (
	"fmt"

	"nimona.io/pkg/context"
	"nimona.io/pkg/mesh"
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

// nolint: lll
type (
	DelegationRequest struct {
		Metadata                object.Metadata         `nimona:"@metadata:m,type=keystream.DelegationRequest"`
		InitiatorConnectionInfo *peer.ConnectionInfo    `nimona:"initiatorConnectionInfo:m"`
		RequestVendor           DelegationRequestVendor `nimona:"requestVendor:m"`
		RequestedPermissions    Permissions             `nimona:"requestedPermissions:m"`
		Nonce                   string                  `nimona:"nonce:s"`
	}
	DelegationRequestVendor struct {
		VendorName             string `nimona:"vendorName:s"`
		VendorURL              string `nimona:"vendorURL:s"`
		ApplicationName        string `nimona:"applicationName:s"`
		ApplicationDescription string `nimona:"applicationDescription:s"`
		ApplicationURL         string `nimona:"applicationURL:s"`
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
	vendor DelegationRequestVendor,
	permissions Permissions,
) (*DelegationRequest, chan Controller, error) {
	ck := m.mesh.GetPeerKey()
	ci := m.mesh.GetConnectionInfo()
	dr := &DelegationRequest{
		Metadata: object.Metadata{
			Owner: ck.PublicKey().DID(),
			// Timestamp: time.Now().Format(time.RFC3339),
		},
		// Nonce:                   rand.String(16),
		InitiatorConnectionInfo: ci,
		RequestVendor:           vendor,
		RequestedPermissions:    permissions,
	}

	drObj, err := object.Marshal(dr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal delegation request: %w", err)
	}

	err = object.Sign(m.mesh.GetPeerKey(), drObj)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign delegation request: %w", err)
	}

	err = m.objectStore.Put(drObj)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to store delegation request: %w", err)
	}

	res := make(chan Controller)

	go func() {
		defer close(res)
		// wait for a Delegation Offer
		env, err := m.mesh.SubscribeOnce(
			ctx,
			mesh.FilterByObjectType("keystream.DelegationOffer"),
		)
		if err != nil {
			// TODO: handle error
			return
		}
		do := &DelegationOffer{}
		err = object.Unmarshal(env.Payload, do)
		if err != nil {
			// TODO: handle error
			return
		}
		// verify Offer
		// TODO compare offer and request
		// if dr.RequestedPermissions != do.DelegatorSeal.Permissions {
		// TODO: handle error
		// return
		// }
		// create new KeyStream
		ctrl, err := m.NewController(&do.DelegatorSeal)
		if err != nil {
			// TODO: handle error
			return
		}
		// send back the root hash of the keystream
		ver := &DelegationVerification{
			Metadata: object.Metadata{
				Owner: ci.PublicKey.DID(),
			},
			DelegateSeal: DelegateSeal{
				Root:        ctrl.GetKeyStream().Root,
				Permissions: permissions,
			},
		}
		verObj, err := object.Marshal(ver)
		if err != nil {
			// TODO: handle error
			return
		}
		err = m.objectStore.Put(verObj)
		if err != nil {
			// TODO: handle error
			return
		}
		err = m.mesh.Send(
			ctx,
			verObj,
			env.Sender,
		)
		if err != nil {
			// TODO: handle error
			return
		}
		// and pass the controller to the caller
		res <- ctrl
	}()

	return dr, res, nil
}

// HandleDelegationRequest handles a DelegationRequest object.
//
// Given a Delegation Request object, creates a DelegationOffer, sends it to
// the Initiator peer, waits for the DelegationVerification, creates a
// Delegation Interaction event and stores it in the controller.
func (m *manager) HandleDelegationRequest(
	ctx context.Context,
	dr *DelegationRequest,
	ks Controller,
) error {
	// create a new Delegation Offer and send it to the
	do := &DelegationOffer{
		Metadata: object.Metadata{
			Owner: ks.GetKeyStream().GetDID(),
		},
		DelegatorSeal: DelegatorSeal{
			Root:        ks.GetKeyStream().Root,
			Sequence:    ks.GetKeyStream().Sequence + 1,
			Permissions: dr.RequestedPermissions,
		},
	}
	doObj, err := object.Marshal(do)
	if err != nil {
		return fmt.Errorf("failed to marshal DelegationOffer: %w", err)
	}
	err = m.mesh.Send(
		ctx,
		doObj,
		dr.InitiatorConnectionInfo.PublicKey,
		mesh.SendWithConnectionInfo(
			dr.InitiatorConnectionInfo,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to send DelegationOffer: %w", err)
	}

	// wait for the DelegationVerification
	env, err := m.mesh.SubscribeOnce(
		ctx,
		mesh.FilterByObjectType("keystream.DelegationVerification"),
	)
	if err != nil {
		return fmt.Errorf("failed to wait for DelegationVerification: %w", err)
	}
	dv := &DelegationVerification{}
	err = object.Unmarshal(env.Payload, dv)
	if err != nil {
		return fmt.Errorf("failed to unmarshal DelegationVerification: %w", err)
	}

	// TODO: verify the DelegationVerification object
	// TODO: we promised a specific sequence number, do we need to check it?

	_, err = ks.Delegate(dv.DelegateSeal)
	if err != nil {
		return fmt.Errorf("failed to create DelegationInteraction: %w", err)
	}

	return nil
}
