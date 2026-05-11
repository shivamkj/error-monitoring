package processing

import "errors"

var (
	ErrInvalidEnvelope   = errors.New("invalid envelope format")
	ErrNoEventInEnvelope = errors.New("no event found in envelope")
)
