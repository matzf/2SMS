package types

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

type BootstrapData struct {
	IA           addr.IA         `json:"ia,omitempty"`
	RawSignature common.RawBytes `json:"raw_signature,omitempty"`
}
