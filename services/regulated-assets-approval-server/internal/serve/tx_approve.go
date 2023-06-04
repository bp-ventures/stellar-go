package serve

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
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

// validate performs some validations on the provided handler data.
func (h txApproveHandler) validate() error {
	if h.issuerKP == nil {
		return errors.New("issuer keypair cannot be nil")
	}
	if h.assetCode == "" {
		return errors.New("asset code cannot be empty")
	}
	if h.horizonClient == nil {
		return errors.New("horizon client cannot be nil")
	}
	if h.networkPassphrase == "" {
		return errors.New("network passphrase cannot be empty")
	}
	if h.db == nil {
		return errors.New("database cannot be nil")
	}
	if h.kycThreshold <= 0 {
		return errors.New("kyc threshold cannot be less than or equal to zero")
	}
	if h.baseURL == "" {
		return errors.New("base url cannot be empty")
	}
	return nil
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
	txSuccessResp, middleOp, err := h.handleSuccessResponseIfNeeded(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "checking if transaction in request was compliant")
	}
	if txSuccessResp != nil {
		return txSuccessResp, nil
	}

	// If we got here it means we have to revise the transaction, that is,
	// inspect the operation and make sure it meets the prerequisites (amount,
	// kyc, etc) for us to authorize it. If prerequisites are met, we take that
	// operation and wrap into SetTrustLineFlags to preserve regulation after
	// the transaction is completed. This is described here:
	// https://github.com/stellar/stellar-protocol/blob/master/ecosystem/sep-0008.md#revised
	// Overview of the revision process, as steps:
	//   1. check if there's only one operation in the tx
	//   2. check if the operation is supported
	//   3. check if asset in the operation is our regulated asset
	//   4. check if account exists in blockchain and tx has correct sequence number
	//   5. check is there's any pending KYC for the source account
	//   6. build a revised transaction, containing our extra operations for regulation
	//   7. sign and return the revised transaction

	// validate the revisable transaction has one operation.
	if len(tx.Operations()) != 1 {
		return NewRejectedTxApprovalResponse("Please submit a transaction with exactly one operation of type payment."), nil
	}
	middleOp = extractMiddleOperation(tx)
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

// validateInput validates if the input parameters contain a valid transaction
// and if the source account is not set in a way that would harm the issuer.
func (h txApproveHandler) validateInput(ctx context.Context, in txApproveRequest) (*txApprovalResponse, *txnbuild.Transaction) {
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

	if tx.SourceAccount().AccountID == h.issuerKP.Address() {
		log.Ctx(ctx).Errorf("transaction sourceAccount is the same as the server issuer account %s", h.issuerKP.Address())
		return NewRejectedTxApprovalResponse("Transaction source account is invalid."), nil
	}

	// only AllowTrust operations can have the issuer as their source account
	for _, op := range tx.Operations() {
		if _, ok := op.(*txnbuild.AllowTrust); ok {
			continue
		}

		if op.GetSourceAccount() == h.issuerKP.Address() {
			log.Ctx(ctx).Error("transaction contains one or more unauthorized operations where source account is the issuer account")
			return NewRejectedTxApprovalResponse("There are one or more unauthorized operations in the provided transaction."), nil
		}
	}

	return nil, tx
}

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

// handleActionRequiredResponseIfNeeded validates and returns an action_required
// response if the payment requires KYC.
func (h txApproveHandler) handleActionRequiredResponseIfNeeded(
	ctx context.Context,
	middleOp *MiddleOperation,
) (*txApprovalResponse, error) {
	amountInt64, err := h.parseAmount(middleOp)
	if err != nil {
		return nil, err
	}
	if amountInt64 <= h.kycThreshold {
		return nil, nil
	}

	intendedCallbackID := uuid.New().String()
	const q = `
		WITH new_row AS (
			INSERT INTO accounts_kyc_status (stellar_address, callback_id)
			VALUES ($1, $2)
			ON CONFLICT(stellar_address) DO NOTHING
			RETURNING *
		)
		SELECT callback_id, approved_at, rejected_at, pending_at FROM new_row
		UNION
		SELECT callback_id, approved_at, rejected_at, pending_at
		FROM accounts_kyc_status
		WHERE stellar_address = $1
	`
	var (
		callbackID                        string
		approvedAt, rejectedAt, pendingAt sql.NullTime
	)
	err = h.db.QueryRowContext(ctx, q, stellarAddress, intendedCallbackID).Scan(&callbackID, &approvedAt, &rejectedAt, &pendingAt)
	if err != nil {
		return nil, errors.Wrap(err, "inserting new row into accounts_kyc_status table")
	}

	if approvedAt.Valid {
		return nil, nil
	}

	kycThreshold, err := convertAmountToReadableString(h.kycThreshold)
	if err != nil {
		return nil, errors.Wrap(err, "converting kycThreshold to human readable string")
	}

	if rejectedAt.Valid {
		return NewRejectedTxApprovalResponse(fmt.Sprintf("Your KYC was rejected and you're not authorized for operations above %s %s.", kycThreshold, h.assetCode)), nil
	}

	if pendingAt.Valid {
		return NewPendingTxApprovalResponse(fmt.Sprintf("Your account could not be verified as approved nor rejected and was marked as pending. You will need staff authorization for operations above %s %s.", kycThreshold, h.assetCode)), nil
	}

	return NewActionRequiredTxApprovalResponse(
		fmt.Sprintf(`Payments exceeding %s %s require KYC approval. Please provide an email address.`, kycThreshold, h.assetCode),
		fmt.Sprintf("%s/kyc-status/%s", h.baseURL, callbackID),
		[]string{"email_address"},
	), nil
}

// handleSuccessResponseIfNeeded inspects the incoming transaction and returns a
// "success" response if it's already compliant with the SEP-8 authorization spec.
func (h txApproveHandler) handleSuccessResponseIfNeeded(ctx context.Context, tx *txnbuild.Transaction) (*txApprovalResponse, *MiddleOperation, error) {
	if len(tx.Operations()) != 5 {
		return nil, nil, nil
	}

	rejectedResp, middleOp := h.validateTransactionOperationsForSuccess(ctx, tx)
	if rejectedResp != nil {
		return rejectedResp, nil, nil
	}

	if middleOp.Payment != nil && middleOp.Payment.Destination == h.issuerKP.Address() {
		return NewRejectedTxApprovalResponse("Can't transfer asset to its issuer."), nil, nil
	}

	// pull current account details from the network then validate the tx sequence number
	acc, err := h.horizonClient.AccountDetail(horizonclient.AccountRequest{AccountID: middleOp.SourceAccount})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "getting detail for payment source account %s", middleOp.SourceAccount)
	}
	if tx.SourceAccount().Sequence != acc.Sequence+1 {
		log.Ctx(ctx).Errorf(`invalid transaction sequence number tx.SourceAccount().Sequence: %d, accountSequence+1: %d`, tx.SourceAccount().Sequence, acc.Sequence+1)
		return NewRejectedTxApprovalResponse("Invalid transaction sequence number."), nil, nil
	}

	kycRequiredResponse, err := h.handleActionRequiredResponseIfNeeded(ctx, middleOp)
	if err != nil {
		return nil, nil, errors.Wrap(err, "handling KYC required payment")
	}
	if kycRequiredResponse != nil {
		return kycRequiredResponse, nil, nil
	}

	// sign transaction with issuer's signature and encode it
	tx, err = tx.Sign(h.networkPassphrase, h.issuerKP)
	if err != nil {
		return nil, nil, errors.Wrap(err, "signing transaction")
	}
	txe, err := tx.Base64()
	if err != nil {
		return nil, nil, errors.Wrap(err, "encoding revised transaction")
	}

	return NewSuccessTxApprovalResponse(txe, "Transaction is compliant and signed by the issuer."), middleOp, nil
}

type MiddleOperation struct {
	SourceAccount   string
	Payment         *txnbuild.Payment
	ManageSellOffer *txnbuild.ManageSellOffer
	ManageBuyOffer  *txnbuild.ManageBuyOffer
}

func extractSourceAccount(sourceAccount string, tx *txnbuild.Transaction) string {
	if sourceAccount != "" {
		return sourceAccount
	}
	return tx.SourceAccount().AccountID
}

// Extract the main operation (which can be a payment, offer or path payment)
// and source account from a transaction.
func extractMiddleOperation(tx *txnbuild.Transaction) *MiddleOperation {
	var opIndex int
	switch len(tx.Operations()) {
	case 1:
		opIndex = 0
	case 3:
		// tx contains an offer operation wrapped by one allow and one disallow
		// trust operations
		opIndex = 1
	case 5:
		// tx contains payment operation wrapped by two allow and two disallow
		// trust operations
		opIndex = 2
	default:
		return nil
	}

	operation := tx.Operations()[opIndex]

	manageSellOfferOp, _ := operation.(*txnbuild.ManageSellOffer)
	if manageSellOfferOp != nil && opIndex == 3 {
		return &MiddleOperation{
			SourceAccount:   extractSourceAccount(manageSellOfferOp.SourceAccount, tx),
			ManageSellOffer: manageSellOfferOp,
		}
	}

	manageBuyOfferOp, _ := operation.(*txnbuild.ManageBuyOffer)
	if manageBuyOfferOp != nil && opIndex == 3 {
		return &MiddleOperation{
			SourceAccount:  extractSourceAccount(manageBuyOfferOp.SourceAccount, tx),
			ManageBuyOffer: manageBuyOfferOp,
		}
	}

	paymentOp, _ := operation.(*txnbuild.Payment)
	if paymentOp != nil && opIndex == 5 {
		return &MiddleOperation{
			SourceAccount: extractSourceAccount(paymentOp.SourceAccount, tx),
			Payment:       paymentOp,
		}
	}

	return nil
}

func containsTrustLineFlags(flags []txnbuild.TrustLineFlag, requiredFlags []txnbuild.TrustLineFlag) bool {
	for _, requiredFlag := range requiredFlags {
		contains := false
		for _, flag := range flags {
			if requiredFlag == flag {
				contains = true
				break
			}
		}
		if !contains {
			return false
		}
	}
	return true
}

func (h txApproveHandler) containsRegulatedAsset(
	middleOp *MiddleOperation,
) bool {
	issuerAddress := h.issuerKP.Address()
	if middleOp.ManageSellOffer != nil {
		offer := middleOp.ManageSellOffer
		return (offer.Selling.GetCode() == h.assetCode &&
			offer.Selling.GetIssuer() == issuerAddress) ||
			(offer.Buying.GetCode() == h.assetCode &&
				offer.Buying.GetIssuer() == issuerAddress)
	} else if middleOp.ManageBuyOffer != nil {
		offer := middleOp.ManageBuyOffer
		return (offer.Selling.GetCode() == h.assetCode &&
			offer.Selling.GetIssuer() == issuerAddress) ||
			(offer.Buying.GetCode() == h.assetCode &&
				offer.Buying.GetIssuer() == issuerAddress)
	} else if middleOp.Payment != nil {
		payment := middleOp.Payment
		return payment.Asset.GetCode() == h.assetCode &&
			payment.Asset.GetIssuer() == issuerAddress
	}
	return false
}

func (h txApproveHandler) operationsValidManageOffer(
	tx *txnbuild.Transaction,
	middleOp *MiddleOperation,
) bool {
	issuerAddress := h.issuerKP.Address()
	op0, ok := tx.Operations()[0].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op0.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op0.SetFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			txnbuild.TrustLineAuthorizedToMaintainLiabilities,
		}) ||
		op0.Trustor != middleOp.SourceAccount ||
		op0.Asset.GetIssuer() != issuerAddress {
		return false
	}
	_, ok = tx.Operations()[1].(*txnbuild.ManageSellOffer)
	if !ok {
		//TODO check price slippage against database
		//TODO check amount
		return false
	}
	op2, ok := tx.Operations()[2].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op2.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op2.ClearFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
		}) ||
		op2.Trustor != middleOp.SourceAccount ||
		op2.Asset.GetIssuer() != issuerAddress {
		return false
	}
	return true
}

func (h txApproveHandler) operationsValidPayment(tx *txnbuild.Transaction, middleOp *MiddleOperation) bool {
	issuerAddress := h.issuerKP.Address()
	op0, ok := tx.Operations()[0].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op0.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op0.SetFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			txnbuild.TrustLineAuthorizedToMaintainLiabilities,
		}) ||
		op0.Trustor != middleOp.SourceAccount ||
		op0.Asset.GetIssuer() != issuerAddress {
		return false
	}
	op1, ok := tx.Operations()[1].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op1.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op1.SetFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			txnbuild.TrustLineAuthorizedToMaintainLiabilities,
		}) ||
		op1.Trustor != middleOp.Payment.Destination ||
		op1.Asset.GetIssuer() != issuerAddress {
		return false
	}
	_, ok = tx.Operations()[2].(*txnbuild.Payment)
	if !ok {
		//TODO check amount
		return false
	}
	op3, ok := tx.Operations()[3].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op3.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op3.ClearFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			// Here we'd also check for txnbuild.TrustLineAuthorizedToMaintainLiabilities,
			// but this would prevent the account from creating offers.
		}) ||
		op3.Trustor != middleOp.SourceAccount ||
		op3.Asset.GetIssuer() != issuerAddress {
		return false
	}
	op4, ok := tx.Operations()[4].(*txnbuild.SetTrustLineFlags)
	if !ok ||
		op4.SourceAccount != issuerAddress ||
		!containsTrustLineFlags(op4.SetFlags, []txnbuild.TrustLineFlag{
			txnbuild.TrustLineAuthorized,
			// Here we'd also check for txnbuild.TrustLineAuthorizedToMaintainLiabilities,
			// but this would prevent the account from creating offers.
		}) ||
		op4.Trustor != middleOp.Payment.Destination ||
		op4.Asset.GetIssuer() != issuerAddress {
		return false
	}
	return true
}

// validateTransactionOperationsForSuccess checks if the incoming transaction
// operations are compliant with the anchor's SEP-8 policy.
func (h txApproveHandler) validateTransactionOperationsForSuccess(ctx context.Context, tx *txnbuild.Transaction) (*txApprovalResponse, *MiddleOperation) {
	if len(tx.Operations()) != 5 {
		return NewRejectedTxApprovalResponse("Unsupported number of operations."), nil
	}

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

func convertAmountToReadableString(threshold int64) (string, error) {
	amountStr := amount.StringFromInt64(threshold)
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return "", errors.Wrap(err, "converting threshold amount from string to float")
	}
	return fmt.Sprintf("%.2f", amountFloat), nil
}
