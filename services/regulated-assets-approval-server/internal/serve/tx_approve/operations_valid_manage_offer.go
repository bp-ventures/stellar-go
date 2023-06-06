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
		op0.Asset.GetIssuer() != issuerAddress {
		return false
	}
	_, ok = tx.Operations()[1].(*txnbuild.ManageSellOffer)
	if ok {
		//TODO check price slippage against database
		//TODO check amount
	} else {
		_, ok = tx.Operations()[1].(*txnbuild.ManageBuyOffer)
		if ok {
			//TODO check price slippage against database
			//TODO check amount
		} else {
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
		op2.Asset.GetIssuer() != issuerAddress {
		return false
	}
	return true
}
