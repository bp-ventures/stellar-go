package tx_approve

import (
	"context"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/txnbuild"
)

// validateInput validates if the input parameters contain a valid transaction
// and if the source account is not set in a way that would harm the issuer.
func (h TxApprove) validateInput(
	ctx context.Context,
	in txApproveRequest,
) (*txApprovalResponse, *txnbuild.Transaction) {
	if in.Tx == "" {
		log.Ctx(ctx).Error(`request is missing parameter "tx".`)
		return NewRejectedTxApprovalResponse(`Missing parameter "tx".`), nil
	}

	genericTx, err := txnbuild.TransactionFromXDR(in.Tx)
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "parsing transaction xdr"))
		return NewRejectedTxApprovalResponse(`Invalid parameter "tx".`), nil
	}

	tx, ok := genericTx.Transaction()
	if !ok {
		log.Ctx(ctx).Error(`invalid parameter "tx", generic transaction not given.`)
		return NewRejectedTxApprovalResponse(`Invalid parameter "tx".`), nil
	}

	if tx.SourceAccount().AccountID == h.IssuerKP.Address() {
		log.Ctx(ctx).Errorf("transaction sourceAccount is the same as the server issuer account %s", h.IssuerKP.Address())
		return NewRejectedTxApprovalResponse("Transaction source account is invalid."), nil
	}

	// only SetTrustLineFlags operations can have the issuer as their source account
	for _, op := range tx.Operations() {
		if _, ok := op.(*txnbuild.SetTrustLineFlags); ok {
			continue
		}

		if op.GetSourceAccount() == h.IssuerKP.Address() {
			log.Ctx(ctx).Error("transaction contains one or more unauthorized operations where source account is the issuer account")
			return NewRejectedTxApprovalResponse("There are one or more unauthorized operations in the provided transaction."), nil
		}
	}

	return nil, tx
}
