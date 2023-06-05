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

type txApproveHandler struct {
	issuerKP          *keypair.Full
	assetCode         string
	horizonClient     horizonclient.ClientInterface
	networkPassphrase string
	db                *sqlx.DB
	kycThreshold      int64
	baseURL           string
}

type txApproveRequest struct {
	Tx string `json:"tx" form:"tx"`
}

type MiddleOperation struct {
	SourceAccount   string
	Payment         *txnbuild.Payment
	ManageSellOffer *txnbuild.ManageSellOffer
	ManageBuyOffer  *txnbuild.ManageBuyOffer
}

func (h txApproveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
func (h txApproveHandler) txApprove(ctx context.Context, in txApproveRequest) (resp *txApprovalResponse, err error) {
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

	// Check if the transaction is already in a format we can sign and approve.
	// If yes, we don't need to revise it, we just sign and return the signed xdr.
	// This covers the "success" case, described here:
	// https://github.com/stellar/stellar-protocol/blob/master/ecosystem/sep-0008.md#success
	txSuccessResp, err := h.handleSuccessResponseIfNeeded(ctx, tx)
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
	//   4. check if account exists in blockchain and tx has correct sequence number
	//   5. check is there's any pending KYC for the source account
	//   6. build a revised transaction, containing our extra operations for regulation
	//   7. sign and return the revised transaction

	// 1. check if there's only one operation in the tx
	if len(tx.Operations()) != 1 {
		return NewRejectedTxApprovalResponse("Please submit a transaction with exactly one operation of type payment."), nil
	}

	// 2. check if the operation is supported
	middleOp := extractMiddleOperation(tx)
	if middleOp == nil {
		return NewRejectedTxApprovalResponse("Unexpected operation in the transaction. Support operations: Payment, ManageSellOffer, ManageBuyOffer."), nil
	}

	if middleOp.Payment != nil && middleOp.Payment.Destination == h.issuerKP.Address() {
		return NewRejectedTxApprovalResponse("Can't transfer asset to its issuer."), nil
	}

	// validate payment asset is the one supported by the issuer
	issuerAddress := h.issuerKP.Address()
	if paymentOp.Asset.GetCode() != h.assetCode || paymentOp.Asset.GetIssuer() != issuerAddress {
		log.Ctx(ctx).Error(`the payment asset is not supported by this issuer`)
		return NewRejectedTxApprovalResponse("The payment asset is not supported by this issuer."), nil
	}

	acc, err := h.horizonClient.AccountDetail(horizonclient.AccountRequest{AccountID: paymentSource})
	if err != nil {
		return nil, errors.Wrapf(err, "getting detail for payment source account %s", paymentSource)
	}

	// validate the sequence number
	if tx.SourceAccount().Sequence != acc.Sequence+1 {
		log.Ctx(ctx).Errorf(`invalid transaction sequence number tx.SourceAccount().Sequence: %d, accountSequence+1: %d`, tx.SourceAccount().Sequence, acc.Sequence+1)
		return NewRejectedTxApprovalResponse("Invalid transaction sequence number."), nil
	}

	actionRequiredResponse, err := h.handleActionRequiredResponseIfNeeded(ctx, middleOp)
	if err != nil {
		return nil, errors.Wrap(err, "handling KYC required payment")
	}
	if actionRequiredResponse != nil {
		return actionRequiredResponse, nil
	}

	// build the transaction
	revisedOperations := []txnbuild.Operation{
		&txnbuild.AllowTrust{
			Trustor:       paymentSource,
			Type:          paymentOp.Asset,
			Authorize:     true,
			SourceAccount: issuerAddress,
		},
		&txnbuild.AllowTrust{
			Trustor:       paymentOp.Destination,
			Type:          paymentOp.Asset,
			Authorize:     true,
			SourceAccount: issuerAddress,
		},
		paymentOp,
		&txnbuild.AllowTrust{
			Trustor:       paymentOp.Destination,
			Type:          paymentOp.Asset,
			Authorize:     false,
			SourceAccount: issuerAddress,
		},
		&txnbuild.AllowTrust{
			Trustor:       paymentSource,
			Type:          paymentOp.Asset,
			Authorize:     false,
			SourceAccount: issuerAddress,
		},
	}
	revisedTx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &acc,
		IncrementSequenceNum: true,
		Operations:           revisedOperations,
		BaseFee:              300,
		Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
	})
	if err != nil {
		return nil, errors.Wrap(err, "building transaction")
	}

	revisedTx, err = revisedTx.Sign(h.networkPassphrase, h.issuerKP)
	if err != nil {
		return nil, errors.Wrap(err, "signing transaction")
	}
	txe, err := revisedTx.Base64()
	if err != nil {
		return nil, errors.Wrap(err, "encoding revised transaction")
	}

	return NewRevisedTxApprovalResponse(txe), nil
}

func extractSourceAccount(sourceAccount string, tx *txnbuild.Transaction) string {
	if sourceAccount != "" {
		return sourceAccount
	}
	return tx.SourceAccount().AccountID
}

func convertAmountToReadableString(threshold int64) (string, error) {
	amountStr := amount.StringFromInt64(threshold)
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return "", errors.Wrap(err, "converting threshold amount from string to float")
	}
	return fmt.Sprintf("%.2f", amountFloat), nil
}
