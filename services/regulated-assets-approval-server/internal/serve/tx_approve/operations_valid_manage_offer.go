package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) operationsValidManageOffer(
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
		!h.isRegulatedAsset(op0.Asset) {
		return false
	}
	_, ok = tx.Operations()[1].(*txnbuild.ManageSellOffer)
	if !ok {
		_, ok = tx.Operations()[1].(*txnbuild.ManageBuyOffer)
		if !ok {
			return false
		}
	}
	op2, ok := tx.Operations()[2].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op2.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op2.ClearFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
		}) ||
		op2.Trustor != middleOp.SourceAccount ||
		!h.isRegulatedAsset(op2.Asset) {
		return false
	}
	return true
}
