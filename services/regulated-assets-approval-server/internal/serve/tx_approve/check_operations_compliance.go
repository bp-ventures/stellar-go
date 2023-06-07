package tx_approve

import (
	"context"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/txnbuild"
)

// checkOperationsCompliance checks if the incoming transaction
// operations are compliant with the anchor's SEP-8 policy.
func (h TxApprove) checkOperationsCompliance(
	ctx context.Context,
	tx *txnbuild.Transaction,
) (*txApprovalResponse, *MiddleOperation) {
	if len(tx.Operations()) != 3 && len(tx.Operations()) != 5 {
		return nil, nil
	}

	middleOp := extractMiddleOperation(tx)
	if middleOp == nil {
		log.Ctx(ctx).Error("no supported middle operation has been found")
		return NewRejectedTxApprovalResponse("There are one or more unexpected operations in the provided transaction."), nil
	}

	if (middleOp.Payment != nil && middleOp.Payment.Destination == h.IssuerKP.Address()) ||
		(middleOp.PathPaymentStrictReceive != nil &&
			middleOp.PathPaymentStrictReceive.Destination == h.IssuerKP.Address()) ||
		(middleOp.PathPaymentStrictSend != nil &&
			middleOp.PathPaymentStrictSend.Destination == h.IssuerKP.Address()) {
		return NewRejectedTxApprovalResponse("Can't transfer asset to its issuer."), nil
	}
	if _, err := h.getRegulatedAsset(middleOp); err != nil {
		return NewRejectedTxApprovalResponse("This asset is not supported by this issuer."), nil
	}

	operationsValid := func() bool {
		if middleOp.Payment != nil {
			return h.operationsValidPayment(tx, middleOp)
		} else if middleOp.PathPaymentStrictReceive != nil {
			return h.operationsValidPathPaymentStrictReceive(tx, middleOp)
		} else if middleOp.PathPaymentStrictSend != nil {
			return h.operationsValidPathPaymentStrictSend(tx, middleOp)
		} else if middleOp.ManageSellOffer != nil {
			return h.operationsValidManageOffer(tx, middleOp)
		} else if middleOp.ManageBuyOffer != nil {
			return h.operationsValidManageOffer(tx, middleOp)
		}
		return true
	}()
	if !operationsValid {
		return NewRejectedTxApprovalResponse("There are one or more unexpected operations in the provided transaction."), nil
	}

	return nil, middleOp
}
