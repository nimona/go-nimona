package keystream

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

type (
	dummyManager struct{}
)

func NewDummyKeyManager() Manager {
	return &dummyManager{}
}

// NewController creates a new controller in the manager's objectstore
func (m *dummyManager) NewController(
	delegatorSeal *DelegatorSeal,
) (Controller, error) {
	return nil, errors.Error("not implemented")
}

func (m *dummyManager) GetController() (Controller, error) {
	return nil, ErrControllerNotFound
}

func (m *dummyManager) WaitForController(
	ctx context.Context,
) (Controller, error) {
	return nil, errors.Error("not implemented")
}

func (m *dummyManager) NewDelegationRequest(
	ctx context.Context,
	vendor DelegationRequestVendor,
	permissions Permissions,
) (*DelegationRequest, chan Controller, error) {
	return nil, nil, errors.Error("not implemented")
}

func (m *dummyManager) HandleDelegationRequest(
	ctx context.Context,
	dr *DelegationRequest,
) error {
	return errors.Error("not implemented")
}
