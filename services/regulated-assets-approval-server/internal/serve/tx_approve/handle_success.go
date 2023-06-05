package tx_approve

import (
	"context"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/txnbuild"
)

// handleSuccessResponseIfNeeded inspects the incoming transaction and returns a
// "success" response if it's already compliant with the SEP-8 authorization spec.
func (h txApproveHandler) handleSuccessResponseIfNeeded(ctx context.Context, tx *txnbuild.Transaction) (*txApprovalResponse, error) {
	if len(tx.Operations()) != 5 {
		return nil, nil
	}

	rejectedResp, middleOp := h.validateTransactionOperationsForSuccess(ctx, tx)
	if rejectedResp != nil {
		return rejectedResp, nil
	}

	if middleOp.Payment != nil && middleOp.Payment.Destination == h.issuerKP.Address() {
		return NewRejectedTxApprovalResponse("Can't transfer asset to its issuer."), nil
	}

	// pull current account details from the network then validate the tx sequence number
	acc, err := h.horizonClient.AccountDetail(horizonclient.AccountRequest{AccountID: middleOp.SourceAccount})
	if err != nil {
		return nil, errors.Wrapf(err, "getting detail for payment source account %s", middleOp.SourceAccount)
	}
	if tx.SourceAccount().Sequence != acc.Sequence+1 {
		log.Ctx(ctx).Errorf(`invalid transaction sequence number tx.SourceAccount().Sequence: %d, accountSequence+1: %d`, tx.SourceAccount().Sequence, acc.Sequence+1)
		return NewRejectedTxApprovalResponse("Invalid transaction sequence number."), nil
	}

	kycRequiredResponse, err := h.handleActionRequiredResponseIfNeeded(ctx, middleOp)
	if err != nil {
		return nil, errors.Wrap(err, "handling KYC required payment")
	}
	if kycRequiredResponse != nil {
		return kycRequiredResponse, nil
	}

	// sign transaction with issuer's signature and encode it
	tx, err = tx.Sign(h.networkPassphrase, h.issuerKP)
	if err != nil {
		return nil, errors.Wrap(err, "signing transaction")
	}
	txe, err := tx.Base64()
	if err != nil {
		return nil, errors.Wrap(err, "encoding revised transaction")
	}

	return NewSuccessTxApprovalResponse(txe, "Transaction is compliant and signed by the issuer."), nil
}
