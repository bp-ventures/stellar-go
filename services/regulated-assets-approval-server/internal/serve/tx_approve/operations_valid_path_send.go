package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) operationsValidPathPaymentStrictSend(
	tx *txnbuild.Transaction,
	middleOp *MiddleOperation,
) bool {
	issuerAddress := h.IssuerKP.Address()
	op0, ok := tx.Operations()[0].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op0.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op0.SetFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
		}) ||
		op0.Trustor != middleOp.SourceAccount ||
		!h.isRegulatedAsset(op0.Asset) {
		return false
	}
	op1, ok := tx.Operations()[1].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op1.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op1.SetFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
		}) ||
		op1.Trustor != middleOp.PathPaymentStrictSend.Destination ||
		!h.isRegulatedAsset(op1.Asset) {
		return false
	}

	_, ok = tx.Operations()[2].(*txnbuild.PathPaymentStrictSend)
	if !ok {
		return false
	}

	op3, ok := tx.Operations()[3].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op3.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op3.ClearFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
		}) ||
		op3.Trustor != middleOp.SourceAccount ||
		!h.isRegulatedAsset(op3.Asset) {
		return false
	}
	op4, ok := tx.Operations()[4].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op4.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op4.ClearFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
		}) ||
		op4.Trustor != middleOp.PathPaymentStrictSend.Destination ||
		!h.isRegulatedAsset(op4.Asset) {
		return false
	}
	return true
}
