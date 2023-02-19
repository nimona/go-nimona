package nimona

import (
	"context"
	"fmt"
)

var ErrDocumentNotFound = fmt.Errorf("document not found")

type (
	DocumentRequest struct {
		_          string     `nimona:"$type,type=core/document.request"`
		Metadata   Metadata   `nimona:"$metadata,omitempty"`
		DocumentID DocumentID `nimona:"documentID"`
	}
	DocumentResponse struct {
		_                string      `nimona:"$type,type=core/document.response"`
		Metadata         Metadata    `nimona:"$metadata,omitempty"`
		Document         DocumentMap `nimona:"document"`
		Found            bool        `nimona:"found"`
		Error            bool        `nimona:"error,omitempty"`
		ErrorDescription string      `nimona:"errorDescription,omitempty"`
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
) (*DocumentMap, error) {
	req := &DocumentRequest{
		Metadata: Metadata{
			Owner: peerConfig.GetIdentity(),
		},
		DocumentID: docID,
	}

	req.Metadata.Signature = NewDocumentSignature(
		peerConfig.GetPrivateKey(),
		NewDocumentHash(req.DocumentMap()),
	)

	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}

	res := &DocumentResponse{}
	err = res.FromDocumentMap(msgRes.DocumentMap)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}

	if !res.Found {
		return nil, fmt.Errorf("got error: %w", ErrDocumentNotFound)
	}

	if res.ErrorDescription != "" {
		return nil, fmt.Errorf("got error: %s", res.ErrorDescription)
	}

	return &res.Document, nil
}

func (h *HandlerDocument) HandleDocumentRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := &DocumentRequest{}
	err := req.FromDocumentMap(msg.DocumentMap)
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

	doc, err := h.DocumentStore.GetDocument(req.DocumentID)
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	if doc == nil {
		return respondWithError("document not found")
	}

	res := &DocumentResponse{
		Found:    true,
		Document: doc.DocumentMap(),
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}
