package tx_approve

import (
	"context"
	"github.com/stellar/go/support/log"
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
		log.Ctx(ctx).Debug("operation is PathPaymentStrictReceive")
		sendMaxFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictReceive.SendMax, 64)
		if err != nil {
			return 0, err
		}
		log.Ctx(ctx).Debugf("sendMaxFloat64: %f", sendMaxFloat64)
		destAmountFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictReceive.DestAmount, 64)
		if err != nil {
			return 0, err
		}
		log.Ctx(ctx).Debugf("destAmountFloat64: %f", destAmountFloat64)
		dbSendUsdRate, err := h.scanUsdRate(ctx, query, middleOp.PathPaymentStrictReceive.SendAsset)
		if err != nil {
			return 0, err
		}
		log.Ctx(ctx).Debugf("dbSendUsdRate: %f", dbSendUsdRate)
		dbDestUsdRate, err := h.scanUsdRate(ctx, query, middleOp.PathPaymentStrictReceive.DestAsset)
		if err != nil {
			return 0, err
		}
		log.Ctx(ctx).Debugf("dbDestUsdRate: %f", dbDestUsdRate)
		opRate := destAmountFloat64 / sendMaxFloat64
		log.Ctx(ctx).Debugf("opRate: %f", opRate)
		dbRate := dbDestUsdRate / dbSendUsdRate
		log.Ctx(ctx).Debugf("dbRate: %f", dbRate)
		diff := math.Abs(opRate - dbRate)
		log.Ctx(ctx).Debugf("diff: %f", diff)
		percDiff := diff / dbRate * 100
		log.Ctx(ctx).Debugf("percDiff: %f", percDiff)
		return percDiff, nil
	} else if middleOp.PathPaymentStrictSend != nil {
		//TODO
		return 0, nil
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
	var usdRate string
	err := h.Db.QueryRowContext(
		ctx,
		query,
		asset.GetCode(),
		asset.GetIssuer(),
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
