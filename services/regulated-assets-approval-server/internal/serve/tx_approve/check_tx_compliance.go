package tx_approve

import (
	"context"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/txnbuild"
)

// checkTxCompliance inspects the incoming transaction and returns a
// "success" response if it's already compliant with the SEP-8 authorization spec.
// Overview of steps performed by this function:
//  1. check if operations match the correct format
//  2. poll account from blockchain and check sequence number
//  3. check if KYC is required for the source account
//  4. sign and return transaction with issuer's signature
func (h TxApprove) checkTxCompliance(
	ctx context.Context,
	successTx *txnbuild.Transaction,
) (*txApprovalResponse, error) {
	// 1. check if operations match the correct format
	rejectedResp, middleOp := h.checkOperationsCompliance(ctx, successTx)
	if rejectedResp != nil {
		return rejectedResp, nil
	}

	// 2. poll account from blockchain and check sequence number
	acc, err := h.HorizonClient.AccountDetail(horizonclient.AccountRequest{AccountID: middleOp.SourceAccount})
	if err != nil {
		return nil, errors.Wrapf(err, "getting detail for payment source account %s", middleOp.SourceAccount)
	}
	if successTx.SourceAccount().Sequence != acc.Sequence+1 {
		log.Ctx(ctx).Errorf(`invalid transaction sequence number tx.SourceAccount().Sequence: %d, accountSequence+1: %d`, successTx.SourceAccount().Sequence, acc.Sequence+1)
		return NewRejectedTxApprovalResponse("Invalid transaction sequence number."), nil
	}

	// 3. check if KYC is required for the source account
	kycRequiredResponse, err := h.checkKyc(ctx, middleOp)
	if err != nil {
		return nil, errors.Wrap(err, "handling KYC required payment")
	}
	if kycRequiredResponse != nil {
		return kycRequiredResponse, nil
	}

	// 4. sign and return transaction with issuer's signature
	successTx, err = successTx.Sign(h.NetworkPassphrase, h.IssuerKP)
	if err != nil {
		return nil, errors.Wrap(err, "signing transaction")
	}
	successTxXdr, err := successTx.Base64()
	if err != nil {
		return nil, errors.Wrap(err, "encoding revised transaction")
	}

	return NewSuccessTxApprovalResponse(successTxXdr, "Transaction is compliant and signed by the issuer."), nil
}
