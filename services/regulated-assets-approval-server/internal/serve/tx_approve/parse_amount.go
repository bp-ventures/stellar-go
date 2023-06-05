package tx_approve

import (
	"github.com/stellar/go/amount"
)

func (h txApproveHandler) parseAmount(middleOp *MiddleOperation) (int64, error) {
	issuerAddress := h.issuerKP.Address()
	var amountInt64 int64
	var err error
	if middleOp.ManageSellOffer != nil {
		if middleOp.ManageSellOffer.Selling.GetIssuer() == issuerAddress {
			amountInt64, err = amount.ParseInt64(middleOp.ManageSellOffer.Amount)
		} else {
			amountInt64, err = amount.ParseInt64(middleOp.ManageSellOffer.Amount)
		}
	} else if middleOp.ManageBuyOffer != nil {
		amountInt64, err = amount.ParseInt64(middleOp.ManageSellOffer.Amount)
	} else if middleOp.Payment != nil {
		amountInt64, err = amount.ParseInt64(paymentOp.Amount)
	}

	if err != nil {
		return 0, err
	}
	return amountInt64
}
