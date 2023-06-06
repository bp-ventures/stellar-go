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
	middleOp := extractMiddleOperation(tx)
	if middleOp == nil {
		log.Ctx(ctx).Error(`middle operation type is not supported`)
		return NewRejectedTxApprovalResponse("There are one or more unexpected operations in the provided transaction."), nil
	}

	if _, err := h.getRegulatedAsset(middleOp); err != nil {
		return NewRejectedTxApprovalResponse("This asset is not supported by this issuer."), nil
	}

	if middleOp.Payment != nil && middleOp.Payment.Destination == h.IssuerKP.Address() {
		return NewRejectedTxApprovalResponse("Can't transfer asset to its issuer."), nil
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
