package tx_approve

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/stellar/go/amount"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/services/regulated-assets-approval-server/internal/serve/httperror"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/support/http/httpdecode"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/txnbuild"
)

type TxApprove struct {
	IssuerKP          *keypair.Full
	AssetCode         string
	HorizonClient     horizonclient.ClientInterface
	NetworkPassphrase string
	Db                *sqlx.DB
	KycThreshold      int64
	BaseURL           string
}

func (h TxApprove) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.validate()
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "validating txApproveHandler"))
		httperror.InternalServer.Render(w)
		return
	}

	in := txApproveRequest{}
	err = httpdecode.Decode(r, &in)
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "decoding txApproveRequest"))
		httperror.BadRequest.Render(w)
		return
	}

	txApproveResp, err := h.txApprove(ctx, in)
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "validating the input transaction for approval"))
		httperror.InternalServer.Render(w)
		return
	}

	txApproveResp.Render(w)
}

// txApprove is called to validate the input transaction.
func (h TxApprove) txApprove(
	ctx context.Context,
	in txApproveRequest,
) (resp *txApprovalResponse, err error) {
	defer func() {
		log.Ctx(ctx).Debug("==== will log responses ====")
		log.Ctx(ctx).Debugf("req: %+v", in)
		log.Ctx(ctx).Debugf("resp: %+v", resp)
		log.Ctx(ctx).Debugf("err: %+v", err)
		log.Ctx(ctx).Debug("====  did log responses ====")
	}()

	rejectedResponse, tx := h.validateInput(ctx, in)
	if rejectedResponse != nil {
		return rejectedResponse, nil
	}

	// Check if the transaction is already compliant, that is, in a format we
	// can sign and approve.
	// If yes, we don't need to revise it, we just sign and return the signed xdr.
	// This covers the "success" case, described here:
	// https://github.com/stellar/stellar-protocol/blob/master/ecosystem/sep-0008.md#success
	log.Ctx(ctx).Debug("Checking if transaction is already compliant")
	txSuccessResp, err := h.checkTxCompliance(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "checking if transaction in request was compliant")
	}
	if txSuccessResp != nil {
		return txSuccessResp, nil
	}

	// If we got here it means we have to revise the transaction, that is,
	// inspect the operation and make sure it meets the prerequisites (amount,
	// kyc, etc) for us to authorize it. If prerequisites are met, we take that
	// operation and wrap into two SetTrustLineFlags operations, where the first
	// one allows the trustline, and the second disallows the trustline.
	// This is described here:
	// https://github.com/stellar/stellar-protocol/blob/master/ecosystem/sep-0008.md#revised
	// Overview of the revision process, as steps:
	//   1. check if there's only one operation in the tx
	//   2. check if the operation is supported
	//   3. check if asset in the operation is our regulated asset
	//   4. poll account from blockchain and check sequence number
	//   5. check if KYC is required for the source account
	//   6. build a revised transaction, containing our extra operations for regulation
	//   7. sign and return the revised transaction

	log.Ctx(ctx).Debug("Revising transaction")

	// 1. check if there's only one operation in the tx
	if len(tx.Operations()) != 1 {
		return NewRejectedTxApprovalResponse("Please submit a transaction with exactly one operation of type payment or offer."), nil
	}

	// 2. check if the operation is supported
	middleOp := extractMiddleOperation(tx)
	if middleOp == nil {
		return NewRejectedTxApprovalResponse("Unexpected operation in the transaction. Support operations: Payment, ManageSellOffer, ManageBuyOffer."), nil
	}

	// 3. check if asset in the operation is our regulated asset
	if middleOp.Payment != nil && middleOp.Payment.Destination == h.IssuerKP.Address() {
		return NewRejectedTxApprovalResponse("Can't transfer asset to its issuer."), nil
	}
	if _, err := h.getRegulatedAsset(middleOp); err != nil {
		log.Ctx(ctx).Error(`asset is not supported by this issuer`)
		return NewRejectedTxApprovalResponse("Asset is not supported by this issuer."), nil
	}

	// 4. poll account from blockchain and check sequence number
	acc, err := h.HorizonClient.AccountDetail(horizonclient.AccountRequest{AccountID: middleOp.SourceAccount})
	if err != nil {
		return nil, errors.Wrapf(err, "getting detail for source account %s", middleOp.SourceAccount)
	}
	if tx.SourceAccount().Sequence != acc.Sequence+1 {
		log.Ctx(ctx).Errorf(`invalid transaction sequence number tx.SourceAccount().Sequence: %d, accountSequence+1: %d`, tx.SourceAccount().Sequence, acc.Sequence+1)
		return NewRejectedTxApprovalResponse("Invalid transaction sequence number."), nil
	}

	// 5. check if KYC is required for the source account
	actionRequiredResponse, err := h.checkKyc(ctx, middleOp)
	if err != nil {
		return nil, errors.Wrap(err, "handling KYC required payment")
	}
	if actionRequiredResponse != nil {
		return actionRequiredResponse, nil
	}

	// 6+7. build and sign revised transaction
	revisedTxXdr, err := h.buildRevisedTx(&acc, middleOp)
	if err != nil {
		return nil, errors.Wrap(err, "building transaction")
	}

	return NewRevisedTxApprovalResponse(revisedTxXdr), nil
}

func extractSourceAccount(sourceAccount string, tx *txnbuild.Transaction) string {
	if sourceAccount != "" {
		return sourceAccount
	}
	return tx.SourceAccount().AccountID
}

func ConvertAmountToReadableString(threshold int64) (string, error) {
	amountStr := amount.StringFromInt64(threshold)
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return "", errors.Wrap(err, "converting threshold amount from string to float")
	}
	return fmt.Sprintf("%.2f", amountFloat), nil
}
