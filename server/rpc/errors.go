package rpc

import "errors"

var (
	ErrFailCast           = errors.New("Failed to cast")
	ErrOverrideMonitoring = errors.New("Override Monitoring")
)
