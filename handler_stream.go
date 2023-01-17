package nimona

import (
	"context"
	"fmt"
)

type (
	StreamRequest struct {
		Type   string     `cborgen:"$type,const=core/stream/request"`
		RootID DocumentID `cbor:"rootDocument"`
	}
	StreamResponse struct {
		Type             string       `cborgen:"$type,const=core/stream/response"`
		RootDocumentID   DocumentID   `cbor:"rootDocument"`
		PatchDocumentIDs []DocumentID `cbor:"patches,omitempty"`
	}
)

type HandlerStream struct {
	DocumentStore *DocumentStore
}

func RequestStream(
	ctx context.Context,
	ses *Session,
	rootID DocumentID,
) (*StreamResponse, error) {
	req := &StreamRequest{
		RootID: rootID,
	}
	res := &StreamResponse{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = msgRes.Decode(res)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerStream) HandleStreamRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := &StreamRequest{}
	err := msg.Decode(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}
	if msg.Type != "core/stream/request" {
		return fmt.Errorf("invalid request type: %s", msg.Type)
	}

	docs, err := h.DocumentStore.GetDocumentsByRootID(req.RootID)
	if err != nil {
		return fmt.Errorf("error getting documents: %w", err)
	}

	var patchIDs []DocumentID
	for _, doc := range docs {
		patchIDs = append(patchIDs, doc.DocumentID)
	}

	res := &StreamResponse{
		RootDocumentID:   req.RootID,
		PatchDocumentIDs: patchIDs,
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}
