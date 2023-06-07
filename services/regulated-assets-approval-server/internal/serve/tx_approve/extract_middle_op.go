package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

// Extract the main operation and source account from a transaction.
func extractMiddleOperation(tx *txnbuild.Transaction) *MiddleOperation {
	var opIndex int
	switch len(tx.Operations()) {
	case 1:
		opIndex = 0
	case 3:
		opIndex = 1
	case 5:
		opIndex = 2
	default:
		return nil
	}

	operation := tx.Operations()[opIndex]

	// Offers are temporarily disabled
	//manageSellOfferOp, _ := operation.(*txnbuild.ManageSellOffer)
	//if manageSellOfferOp != nil {
	//	return &MiddleOperation{
	//		SourceAccount:   extractSourceAccount(manageSellOfferOp.SourceAccount, tx),
	//		ManageSellOffer: manageSellOfferOp,
	//	}
	//}

	//manageBuyOfferOp, _ := operation.(*txnbuild.ManageBuyOffer)
	//if manageBuyOfferOp != nil {
	//	return &MiddleOperation{
	//		SourceAccount:  extractSourceAccount(manageBuyOfferOp.SourceAccount, tx),
	//		ManageBuyOffer: manageBuyOfferOp,
	//	}
	//}

	paymentOp, _ := operation.(*txnbuild.Payment)
	if paymentOp != nil {
		return &MiddleOperation{
			SourceAccount: extractSourceAccount(paymentOp.SourceAccount, tx),
			Payment:       paymentOp,
		}
	}

	pathPaymentStrictReceiveOp, _ := operation.(*txnbuild.PathPaymentStrictReceive)
	if pathPaymentStrictReceiveOp != nil {
		return &MiddleOperation{
			SourceAccount:            extractSourceAccount(pathPaymentStrictReceiveOp.SourceAccount, tx),
			PathPaymentStrictReceive: pathPaymentStrictReceiveOp,
		}
	}

	pathPaymentStrictSendOp, _ := operation.(*txnbuild.PathPaymentStrictSend)
	if pathPaymentStrictSendOp != nil {
		return &MiddleOperation{
			SourceAccount:         extractSourceAccount(pathPaymentStrictSendOp.SourceAccount, tx),
			PathPaymentStrictSend: pathPaymentStrictSendOp,
		}
	}

	return nil
}
