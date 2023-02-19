package nimona

import (
	"context"
	"fmt"
)

type (
	DocumentGraphRequest struct {
		_              string     `nimona:"$type,type=core/document/graph.request"`
		RootDocumentID DocumentID `nimona:"rootDocument"`
	}
	DocumentGraphResponse struct {
		_                string       `nimona:"$type,type=core/document/graph.response"`
		RootDocumentID   DocumentID   `nimona:"rootDocumentID"`
		PatchDocumentIDs []DocumentID `nimona:"patchDocumentIDs"`
	}
)

type HandlerDocumentGraph struct {
	DocumentStore *DocumentStore
}

func RequestDocumentGraph(
	ctx context.Context,
	ses *Session,
	rootID DocumentID,
) (*DocumentGraphResponse, error) {
	req := &DocumentGraphRequest{
		RootDocumentID: rootID,
	}
	res := DocumentGraphResponse{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = res.FromDocumentMap(msgRes.DocumentMap)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return &res, nil
}

func (h *HandlerDocumentGraph) HandleDocumentGraphRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := DocumentGraphRequest{}
	req.FromDocumentMap(msg.DocumentMap)
	if msg.Type != "core/document/graph.request" {
		return fmt.Errorf("invalid request type: %s", msg.Type)
	}

	docs, err := h.DocumentStore.GetDocumentsByRootID(req.RootDocumentID)
	if err != nil {
		return fmt.Errorf("error getting documents: %w", err)
	}

	var patchIDs []DocumentID
	for _, doc := range docs {
		patchIDs = append(patchIDs, doc.DocumentID)
	}

	res := &DocumentGraphResponse{
		RootDocumentID:   req.RootDocumentID,
		PatchDocumentIDs: patchIDs,
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}
