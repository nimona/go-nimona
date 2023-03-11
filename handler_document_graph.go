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

func SyncDocumentGraph(
	ctx context.Context,
	rctx *RequestContext,
	ses *SessionManager,
	rootID DocumentID,
	rec RequestRecipientFn,
) (*Document, []*DocumentPatch, error) {
	graphRes, err := RequestDocumentGraph(
		ctx,
		rctx,
		ses,
		rootID,
		rec,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error requesting graph: %w", err)
	}

	loadOrRequest := func(id DocumentID) (*Document, error) {
		if rctx.DocumentStore != nil {
			doc, loadErr := rctx.DocumentStore.GetDocument(id)
			if loadErr == nil {
				return doc, nil
			}
		}
		doc, reqErr := RequestDocument(ctx, rctx, ses, id, rec)
		if reqErr != nil {
			return nil, fmt.Errorf("error requesting document: %w", err)
		}
		if rctx.DocumentStore != nil {
			rctx.DocumentStore.PutDocument(doc)
		}
		return doc, nil
	}

	// get root document
	rootDoc, err := loadOrRequest(graphRes.RootDocumentID)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading root document: %w", err)
	}

	// get patch documents
	patches := []*DocumentPatch{}
	for _, gotPatchDocID := range graphRes.PatchDocumentIDs {
		gotPatchDoc, err := RequestDocument(ctx, rctx, ses, gotPatchDocID, rec)
		if err != nil {
			return nil, nil, fmt.Errorf("error requesting patch document: %w", err)
		}
		gotPatch := &DocumentPatch{}
		err = gotPatch.FromDocument(gotPatchDoc)
		if err != nil {
			return nil, nil, fmt.Errorf("error decoding patch document: %w", err)
		}
		patches = append(patches, gotPatch)
	}

	return rootDoc, patches, nil
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
			patchIDs = append(patchIDs, NewDocumentID(doc))
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
