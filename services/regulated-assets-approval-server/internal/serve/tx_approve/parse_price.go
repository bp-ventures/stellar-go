package tx_approve

import (
	"context"
	"math"
	"strconv"

	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) getUsdPricePercentageDiff(
	ctx context.Context,
	middleOp *MiddleOperation) (float64, error) {
	const query = `
		SELECT usd_rate
		FROM fx_rates
		WHERE asset_code = $1 AND
			  asset_issuer = $2
		ORDER BY rate_timestamp DESC
	`
	if middleOp.Payment != nil {
		return 0, errors.New("cannot parse price for payment operation")
	} else if middleOp.PathPaymentStrictReceive != nil {
		sendMaxFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictReceive.SendMax, 64)
		if err != nil {
			return 0, err
		}
		destAmountFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictReceive.DestAmount, 64)
		if err != nil {
			return 0, err
		}
		dbSendUsdRate, err := h.scanUsdRate(ctx, query, middleOp.PathPaymentStrictReceive.SendAsset)
		if err != nil {
			return 0, err
		}
		dbDestUsdRate, err := h.scanUsdRate(ctx, query, middleOp.PathPaymentStrictReceive.DestAsset)
		if err != nil {
			return 0, err
		}
		opRate := destAmountFloat64 / sendMaxFloat64
		dbRate := dbDestUsdRate / dbSendUsdRate
		diff := math.Abs(opRate - dbRate)
		percDiff := diff / dbRate * 100
		return percDiff, nil
	} else if middleOp.PathPaymentStrictSend != nil {
		sendAmountFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictSend.SendAmount, 64)
		if err != nil {
			return 0, err
		}
		destMinFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictSend.DestMin, 64)
		if err != nil {
			return 0, err
		}
		dbSendUsdRate, err := h.scanUsdRate(ctx, query, middleOp.PathPaymentStrictSend.SendAsset)
		if err != nil {
			return 0, err
		}
		dbDestUsdRate, err := h.scanUsdRate(ctx, query, middleOp.PathPaymentStrictSend.DestAsset)
		if err != nil {
			return 0, err
		}
		opRate := destMinFloat64 / sendAmountFloat64
		dbRate := dbDestUsdRate / dbSendUsdRate
		diff := math.Abs(opRate - dbRate)
		percDiff := diff / dbRate * 100
		return percDiff, nil
	} else if middleOp.ManageSellOffer != nil {
		//TODO
		return 0, nil
	} else if middleOp.ManageBuyOffer != nil {
		//TODO
		return 0, nil
	} else {
		return 0, errors.New("middleOp has no operation set")
	}
}

func (h TxApprove) scanUsdRate(ctx context.Context, query string, asset txnbuild.Asset) (float64, error) {
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
		return 0, err
	}
	usdRateFloat64, err := strconv.ParseFloat(usdRate, 64)
	if err != nil {
		return 0, err
	}
	return usdRateFloat64, err
}
