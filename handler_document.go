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
		_                string    `nimona:"$type,type=core/document.response"`
		Metadata         Metadata  `nimona:"$metadata,omitempty"`
		Payload          *Document `nimona:"document"`
		Found            bool      `nimona:"found"`
		Error            bool      `nimona:"error,omitempty"`
		ErrorDescription string    `nimona:"errorDescription,omitempty"`
	}
)

func RequestDocument(
	ctx context.Context,
	rctx *RequestContext,
	ses *SessionManager,
	docID DocumentID,
	rec RequestRecipientFn,
) (*Document, error) {
	req := DocumentRequest{
		DocumentID: docID,
	}

	doc := req.Document()
	SignDocument(rctx, doc)

	msgRes, err := ses.Request(ctx, doc, rec)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}

	res := &DocumentResponse{}
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}

	if !res.Found {
		return nil, fmt.Errorf("got error: %w", ErrDocumentNotFound)
	}

	if res.ErrorDescription != "" {
		return nil, fmt.Errorf("got error: %s", res.ErrorDescription)
	}

	return res.Payload, nil
}

func HandleDocumentRequest(
	sesManager *SessionManager,
	docStore *DocumentStore,
) {
	handler := func(
		ctx context.Context,
		msg *Request,
	) error {
		req := &DocumentRequest{}
		err := req.FromDocument(msg.Document)
		if err != nil {
			return fmt.Errorf("document handler - error unmarshaling request: %w", err)
		}

		respondWithError := func(desc string) error {
			res := &DocumentResponse{
				Error:            true,
				ErrorDescription: desc,
			}
			err = msg.Respond(res.Document())
			if err != nil {
				return fmt.Errorf("document handler - error replying: %w", err)
			}
			return nil
		}

		doc, err := docStore.GetDocument(req.DocumentID)
		if err != nil {
			respondWithError("document not found")
			return fmt.Errorf("document handler - error getting document: %w", err)
		}

		if doc == nil {
			return respondWithError("document not found")
		}

		res := &DocumentResponse{
			Found:   true,
			Payload: doc.Document().Copy(),
		}
		err = msg.Respond(res.Document())
		if err != nil {
			return fmt.Errorf("document handler - error replying: %w", err)
		}
		return nil
	}
	sesManager.RegisterHandler("core/document.request", handler)
}

type (
	DocumentStoreRequest struct {
		_        string    `nimona:"$type,type=core/documentStore.request"`
		Metadata Metadata  `nimona:"$metadata,omitempty"`
		Payload  *Document `nimona:"document"`
	}
	DocumentStoreResponse struct {
		_                string   `nimona:"$type,type=core/documentStore.response"`
		Metadata         Metadata `nimona:"$metadata,omitempty"`
		Error            bool     `nimona:"error,omitempty"`
		ErrorDescription string   `nimona:"errorDescription,omitempty"`
	}
)

func PublishDocument(
	ctx context.Context,
	ses *SessionManager,
	rctx *RequestContext,
	payload *Document,
	rec RequestRecipientFn,
) error {
	req := DocumentStoreRequest{
		Payload: payload,
	}

	doc := req.Document()
	SignDocument(rctx, doc)

	msgRes, err := ses.Request(ctx, doc, rec)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	res := &DocumentStoreResponse{}
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return fmt.Errorf("error decoding message: %w", err)
	}

	if res.Error {
		if res.ErrorDescription != "" {
			return fmt.Errorf("received error response: %s", res.ErrorDescription)
		}
		return fmt.Errorf("received error response")
	}

	return nil
}

func HandleDocumentStoreRequest(
	sesManager *SessionManager,
	docStore *DocumentStore,
) {
	handler := func(
		ctx context.Context,
		msg *Request,
	) error {
		req := &DocumentStoreRequest{}
		err := req.FromDocument(msg.Document)
		if err != nil {
			return fmt.Errorf("error unmarshaling request: %w", err)
		}

		respondWithError := func(desc string) error {
			res := &DocumentStoreResponse{
				Error:            true,
				ErrorDescription: desc,
			}
			err = msg.Respond(res.Document())
			if err != nil {
				return fmt.Errorf("error replying: %w", err)
			}
			return nil
		}

		if req.Payload == nil {
			return respondWithError("missing document")
		}

		err = docStore.PutDocument(req.Payload)
		if err != nil {
			return fmt.Errorf("error storing document: %w", err)
		}

		res := &DocumentStoreResponse{}
		err = msg.Respond(res.Document())
		if err != nil {
			return fmt.Errorf("error replying: %w", err)
		}
		return nil
	}
	sesManager.RegisterHandler("core/documentStore.request", handler)
}
