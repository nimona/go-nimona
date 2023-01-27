package nimona

import (
	"context"
	"fmt"

	cbg "github.com/whyrusleeping/cbor-gen"
)

type (
	DocumentRequest struct {
		_          string     `cborgen:"$type,const=core/document.request"`
		Metadata   Metadata   `cborgen:"$metadata,omitempty"`
		DocumentID DocumentID `cborgen:"documentID"`
	}
	DocumentResponse struct {
		_                string       `cborgen:"$type,const=core/document.response"`
		Metadata         Metadata     `cborgen:"$metadata,omitempty"`
		Document         cbg.Deferred `cborgen:"document"`
		Found            bool         `cborgen:"found"`
		Error            bool         `cborgen:"error,omitempty"`
		ErrorDescription string       `cborgen:"errorDescription,omitempty"`
	}
)

type (
	HandlerDocument struct {
		Hostname      string
		PeerConfig    *PeerConfig
		DocumentStore *DocumentStore
	}
)

func RequestDocument(
	ctx context.Context,
	ses *Session,
	peerConfig *PeerConfig,
	docID DocumentID,
) (*DocumentBase, error) {
	req := &DocumentRequest{
		Metadata: Metadata{
			Owner: peerConfig.GetIdentity().IdentityID(),
		},
		DocumentID: docID,
	}
	docHash, err := NewDocumentHash(req)
	if err != nil {
		return nil, fmt.Errorf("error creating document hash: %w", err)
	}

	docSig, err := NewDocumentSignature(
		peerConfig.GetPrivateKey(),
		docHash,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating document signature: %w", err)
	}

	req.Metadata.Signature = *docSig

	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}

	res := &DocumentResponse{}
	err = msgRes.Decode(res)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}

	doc := &DocumentBase{}
	err = UnmarshalCBORBytes(res.Document.Raw, doc)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling document: %w", err)
	}

	return doc, nil
}

func (h *HandlerDocument) HandleDocumentRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := &DocumentRequest{}
	err := msg.Decode(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}

	respondWithError := func(desc string) error {
		res := &DocumentResponse{
			Error:            true,
			ErrorDescription: desc,
		}
		err = msg.Respond(res)
		if err != nil {
			return fmt.Errorf("error replying: %w", err)
		}
		return nil
	}

	docEntry, err := h.DocumentStore.GetDocument(req.DocumentID)
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}
	if docEntry == nil {
		return respondWithError("document not found")
	}

	res := &DocumentResponse{
		Found: true,
		Document: cbg.Deferred{
			Raw: docEntry.DocumentBytes,
		},
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}
