package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) buildRevisedOperationsPayment(
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
			},
		},
		&txnbuild.SetTrustLineFlags{
			SourceAccount: issuerAddress,
			Trustor:       middleOp.Payment.Destination,
			Asset:         asset,
			SetFlags: []txnbuild.TrustLineFlag{
				txnbuild.TrustLineAuthorized,
			},
		},
		middleOp.Payment,
		&txnbuild.SetTrustLineFlags{
			SourceAccount: issuerAddress,
			Trustor:       middleOp.SourceAccount,
			Asset:         asset,
			ClearFlags: []txnbuild.TrustLineFlag{
				txnbuild.TrustLineAuthorized,
			},
		},
		&txnbuild.SetTrustLineFlags{
			SourceAccount: issuerAddress,
			Trustor:       middleOp.Payment.Destination,
			Asset:         asset,
			ClearFlags: []txnbuild.TrustLineFlag{
				txnbuild.TrustLineAuthorized,
			},
		},
	}
}
