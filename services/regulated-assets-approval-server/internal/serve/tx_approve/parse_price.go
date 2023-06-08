package tx_approve

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) getUsdPricePercentageDiff(
	ctx context.Context,
	middleOp *MiddleOperation) (*decimal.Decimal, error) {
	if middleOp.Payment != nil {
		return nil, errors.New("cannot parse price for payment operation")
	} else if middleOp.PathPaymentStrictReceive != nil {
		sendMax, err := decimal.NewFromString(middleOp.PathPaymentStrictReceive.SendMax)
		if err != nil {
			return nil, err
		}
		destAmount, err := decimal.NewFromString(middleOp.PathPaymentStrictReceive.DestAmount)
		if err != nil {
			return nil, err
		}
		dbSendUsdRate, err := h.scanUsdRate(ctx, middleOp.PathPaymentStrictReceive.SendAsset)
		if err != nil {
			return nil, err
		}
		dbDestUsdRate, err := h.scanUsdRate(ctx, middleOp.PathPaymentStrictReceive.DestAsset)
		if err != nil {
			return nil, err
		}
		opRate := destAmount.Div(sendMax)
		dbRate := dbDestUsdRate.Div(*dbSendUsdRate)
		diff := opRate.Sub(dbRate).Abs()
		percDiff := diff.Div(dbRate).Mul(decimal.NewFromInt(100))
		return &percDiff, nil
	} else if middleOp.PathPaymentStrictSend != nil {
		sendAmount, err := decimal.NewFromString(middleOp.PathPaymentStrictSend.SendAmount)
		if err != nil {
			return nil, err
		}
		destMin, err := decimal.NewFromString(middleOp.PathPaymentStrictSend.DestMin)
		if err != nil {
			return nil, err
		}
		dbSendUsdRate, err := h.scanUsdRate(ctx, middleOp.PathPaymentStrictSend.SendAsset)
		if err != nil {
			return nil, err
		}
		dbDestUsdRate, err := h.scanUsdRate(ctx, middleOp.PathPaymentStrictSend.DestAsset)
		if err != nil {
			return nil, err
		}
		opRate := destMin.Div(sendAmount)
		dbRate := dbDestUsdRate.Div(*dbSendUsdRate)
		diff := opRate.Sub(dbRate).Abs()
		percDiff := diff.Div(dbRate).Mul(decimal.NewFromInt(100))
		return &percDiff, nil
	} else if middleOp.ManageSellOffer != nil {
		//TODO
		return nil, nil
	} else if middleOp.ManageBuyOffer != nil {
		//TODO
		return nil, nil
	} else {
		return nil, errors.New("middleOp has no operation set")
	}
}

func (h TxApprove) scanUsdRate(ctx context.Context, asset txnbuild.Asset) (*decimal.Decimal, error) {
	const query = `
		SELECT usd_rate
		FROM fx_rates
		WHERE asset_code = $1 AND
			  asset_issuer = $2
		ORDER BY rate_timestamp DESC
	`
	var (
		usdRate     string
		assetCode   string
		assetIssuer string
	)
	if asset.IsNative() {
		assetCode = "XLM"
		assetIssuer = ""
	} else {
		assetCode = asset.GetCode()
		assetIssuer = asset.GetIssuer()
	}
	err := h.Db.QueryRowContext(
		ctx,
		query,
		assetCode,
		assetIssuer,
	).Scan(&usdRate)
	if err != nil {
		return nil, err
	}
	usdRateDecimal, err := decimal.NewFromString(usdRate)
	if err != nil {
		return nil, err
	}
	return &usdRateDecimal, err
}
