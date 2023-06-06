package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

// Extract the main operation (which can be a payment, offer or path payment)
// and source account from a transaction.
func extractMiddleOperation(tx *txnbuild.Transaction) *MiddleOperation {
	var opIndex int
	switch len(tx.Operations()) {
	case 1:
		opIndex = 0
	case 3:
		// tx contains an offer operation wrapped by one allow and one disallow
		// trust operations
		opIndex = 1
	case 5:
		// tx contains payment operation wrapped by two allow and two disallow
		// trust operations
		opIndex = 2
	default:
		return nil
	}

	operation := tx.Operations()[opIndex]

	manageSellOfferOp, _ := operation.(*txnbuild.ManageSellOffer)
	if manageSellOfferOp != nil {
		return &MiddleOperation{
			SourceAccount:   extractSourceAccount(manageSellOfferOp.SourceAccount, tx),
			ManageSellOffer: manageSellOfferOp,
		}
	}

	manageBuyOfferOp, _ := operation.(*txnbuild.ManageBuyOffer)
	if manageBuyOfferOp != nil {
		return &MiddleOperation{
			SourceAccount:  extractSourceAccount(manageBuyOfferOp.SourceAccount, tx),
			ManageBuyOffer: manageBuyOfferOp,
		}
	}

	paymentOp, _ := operation.(*txnbuild.Payment)
	if paymentOp != nil {
		return &MiddleOperation{
			SourceAccount: extractSourceAccount(paymentOp.SourceAccount, tx),
			Payment:       paymentOp,
		}
	}

	return nil
}
