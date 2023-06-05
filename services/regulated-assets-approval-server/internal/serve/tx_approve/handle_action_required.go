package tx_approve

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/stellar/go/support/errors"
)

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
	err = h.db.QueryRowContext(ctx, q, middleOp.SourceAccount, intendedCallbackID).Scan(&callbackID, &approvedAt, &rejectedAt, &pendingAt)
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
