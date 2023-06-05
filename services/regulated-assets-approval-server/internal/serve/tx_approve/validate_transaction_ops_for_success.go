package tx_approve

import (
	"context"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/txnbuild"
)

// validateTransactionOperationsForSuccess checks if the incoming transaction
// operations are compliant with the anchor's SEP-8 policy.
func (h txApproveHandler) validateTransactionOperationsForSuccess(ctx context.Context, tx *txnbuild.Transaction) (*txApprovalResponse, *MiddleOperation) {
	middleOp := extractMiddleOperation(tx)
	if middleOp == nil {
		log.Ctx(ctx).Error(`middle operation is not payment, offer or path payment`)
		return NewRejectedTxApprovalResponse("There are one or more unexpected operations in the provided transaction."), nil
	}

	if !h.containsAssetIssuer(middleOp) {
		return NewRejectedTxApprovalResponse("This asset is not supported by this issuer."), nil
	}

	operationsValid := func() bool {
		if middleOp.ManageSellOffer != nil {
			return h.operationsValidManageOffer(tx, middleOp)
		} else if middleOp.ManageBuyOffer != nil {
			return h.operationsValidManageOffer(tx, middleOp)
		} else if middleOp.Payment != nil {
			return h.operationsValidPayment(tx, middleOp)
		}
		return true
	}()
	if !operationsValid {
		return NewRejectedTxApprovalResponse("There are one or more unexpected operations in the provided transaction."), nil
	}

	return nil, middleOp
}
