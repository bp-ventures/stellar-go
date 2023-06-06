package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) operationsValidPayment(
	tx *txnbuild.Transaction,
	middleOp *MiddleOperation,
) bool {
	issuerAddress := h.IssuerKP.Address()
	op0, ok := tx.Operations()[0].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op0.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op0.SetFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			txnbuild.TrustLineAuthorizedToMaintainLiabilities,
		}) ||
		op0.Trustor != middleOp.SourceAccount ||
		op0.Asset.GetIssuer() != issuerAddress {
		return false
	}
	op1, ok := tx.Operations()[1].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op1.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op1.SetFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			txnbuild.TrustLineAuthorizedToMaintainLiabilities,
		}) ||
		op1.Trustor != middleOp.Payment.Destination ||
		op1.Asset.GetIssuer() != issuerAddress {
		return false
	}
	_, ok = tx.Operations()[2].(*txnbuild.Payment)
	if !ok {
		//TODO check amount
		return false
	}
	op3, ok := tx.Operations()[3].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op3.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op3.ClearFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			// Here we'd also check for txnbuild.TrustLineAuthorizedToMaintainLiabilities,
			// but this would prevent the account from creating offers.
		}) ||
		op3.Trustor != middleOp.SourceAccount ||
		op3.Asset.GetIssuer() != issuerAddress {
		return false
	}
	op4, ok := tx.Operations()[4].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op4.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op4.ClearFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			// Here we'd also check for txnbuild.TrustLineAuthorizedToMaintainLiabilities,
			// but this would prevent the account from creating offers.
		}) ||
		op4.Trustor != middleOp.Payment.Destination ||
		op4.Asset.GetIssuer() != issuerAddress {
		return false
	}
	return true
}
