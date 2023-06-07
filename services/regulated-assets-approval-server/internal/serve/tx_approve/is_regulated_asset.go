package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) isRegulatedAsset(asset txnbuild.Asset) bool {
	issuerAddress := h.IssuerKP.Address()
	return asset.GetCode() == h.AssetCode && asset.GetIssuer() == issuerAddress
}
