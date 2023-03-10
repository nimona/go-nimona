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

func RequestDocumentGraph(
	ctx context.Context,
	rctx *RequestContext,
	ses *SessionManager,
	rootID DocumentID,
	rec RequestRecipientFn,
) (*DocumentGraphResponse, error) {
	req := &DocumentGraphRequest{
		RootDocumentID: rootID,
	}

	doc := req.Document()
	SignDocument(rctx, doc)

	res := DocumentGraphResponse{}
	msgRes, err := ses.Request(ctx, doc, rec)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return &res, nil
}

func HandleDocumentGraphRequest(
	sesManager *SessionManager,
	docStore *DocumentStore,
) {
	handler := func(ctx context.Context, msg *Request) error {
		if msg.Type != "core/document/graph.request" {
			return fmt.Errorf("invalid request type: %s", msg.Type)
		}

		req := DocumentGraphRequest{}
		err := req.FromDocument(msg.Document)
		if err != nil {
			return fmt.Errorf("error decoding message: %w", err)
		}

		docs, err := docStore.GetDocumentsByRootID(req.RootDocumentID)
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
		err = msg.Respond(res.Document())
		if err != nil {
			return fmt.Errorf("error replying: %w", err)
		}
		return nil
	}
	sesManager.RegisterHandler("core/document/graph.request", handler)
}
