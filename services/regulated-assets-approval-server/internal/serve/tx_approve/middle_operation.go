package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

type MiddleOperation struct {
	SourceAccount            string
	Payment                  *txnbuild.Payment
	ManageSellOffer          *txnbuild.ManageSellOffer
	ManageBuyOffer           *txnbuild.ManageBuyOffer
	PathPaymentStrictReceive *txnbuild.PathPaymentStrictReceive
	PathPaymentStrictSend    *txnbuild.PathPaymentStrictSend
}
