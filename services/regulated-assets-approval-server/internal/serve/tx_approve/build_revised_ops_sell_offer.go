package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) buildRevisedOperationsManageSellOffer(
	acc txnbuild.Account,
	middleOp *MiddleOperation,
	asset txnbuild.Asset,
	issuerAddress string,
) []txnbuild.Operation {
	return []txnbuild.Operation{
		&txnbuild.SetTrustLineFlags{
			SourceAccount: issuerAddress,
			Trustor:       middleOp.SourceAccount,
			Asset:         asset,
			SetFlags: []txnbuild.TrustLineFlag{
				txnbuild.TrustLineAuthorized,
				txnbuild.TrustLineAuthorizedToMaintainLiabilities,
			},
		},
		middleOp.ManageSellOffer,
		&txnbuild.SetTrustLineFlags{
			SourceAccount: issuerAddress,
			Trustor:       middleOp.SourceAccount,
			Asset:         asset,
			ClearFlags: []txnbuild.TrustLineFlag{
				txnbuild.TrustLineAuthorized,
			},
		},
	}
}
