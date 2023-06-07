package tx_approve

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stellar/go/support/log"

	"github.com/google/uuid"
	"github.com/stellar/go/support/errors"
)

// checkKyc validates and returns an action_required
// response if the payment requires KYC.
func (h TxApprove) checkKyc(
	ctx context.Context,
	middleOp *MiddleOperation,
) (*txApprovalResponse, error) {
	//TODO move this check to after the amount check
	if priceRequiresKyc, err := h.priceRequiresKyc(ctx, middleOp); err != nil {
		return nil, err
	} else if !priceRequiresKyc {
		return nil, nil
	}
	if amountRequiresKyc, err := h.amountRequiresKyc(middleOp); err != nil {
		return nil, err
	} else if !amountRequiresKyc {
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
	err := h.Db.QueryRowContext(ctx, q, middleOp.SourceAccount, intendedCallbackID).Scan(&callbackID, &approvedAt, &rejectedAt, &pendingAt)
	if err != nil {
		return nil, errors.Wrap(err, "inserting new row into accounts_kyc_status table")
	}

	if approvedAt.Valid {
		return nil, nil
	}

	kycThreshold, err := ConvertAmountToReadableString(h.KycPaymentThreshold)
	if err != nil {
		return nil, errors.Wrap(err, "converting kycThreshold to human readable string")
	}

	if rejectedAt.Valid {
		return NewRejectedTxApprovalResponse(fmt.Sprintf("Your KYC was rejected and you're not authorized for operations above %s %s.", kycThreshold, h.AssetCode)), nil
	}

	if pendingAt.Valid {
		return NewPendingTxApprovalResponse(fmt.Sprintf("Your account could not be verified as approved nor rejected and was marked as pending. You will need staff authorization for operations above %s %s.", kycThreshold, h.AssetCode)), nil
	}

	return NewActionRequiredTxApprovalResponse(
		fmt.Sprintf(`Payments exceeding %s %s require KYC approval. Please provide an email address.`, kycThreshold, h.AssetCode),
		fmt.Sprintf("%s/kyc-status/%s", h.BaseURL, callbackID),
		[]string{"email_address"},
	), nil
}

func (h TxApprove) amountRequiresKyc(middleOp *MiddleOperation) (bool, error) {
	amountInt64, err := h.parseAmount(middleOp)
	if err != nil {
		return false, err
	}
	if amountInt64 <= h.KycPaymentThreshold {
		return false, nil
	}
	return true, nil
}

func (h TxApprove) priceRequiresKyc(ctx context.Context, middleOp *MiddleOperation) (bool, error) {
	log.Ctx(ctx).Debug("getting usd price percentage diff")
	percDiff, err := h.getUsdPricePercentageDiff(ctx, middleOp)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Ctx(ctx).Debug("no rows found")
			return false, nil
		}
		return false, err
	}
	if percDiff > 5.0 {
		return true, nil
	}
	return false, nil
}
