package tx_approve

import (
	"errors"
	"math"
	"strconv"
)

func (h TxApprove) parseAmount(middleOp *MiddleOperation) (int64, error) {
	if middleOp.Payment != nil {
		if amountFloat64, err := strconv.ParseFloat(middleOp.Payment.Amount, 64); err != nil {
			return 0, err
		} else {
			return int64(math.Ceil(amountFloat64)), nil
		}
	} else if middleOp.PathPaymentStrictReceive != nil {
		if h.isRegulatedAsset(middleOp.PathPaymentStrictReceive.SendAsset) {
			sendMaxFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictReceive.SendMax, 64)
			if err != nil {
				return 0, err
			}
			sendMaxInt64 := int64(math.Ceil(sendMaxFloat64))
			return sendMaxInt64, nil
		} else {
			destAmountFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictReceive.DestAmount, 64)
			if err != nil {
				return 0, err
			}
			destAmountInt64 := int64(math.Ceil(destAmountFloat64))
			return destAmountInt64, nil
		}
	} else if middleOp.PathPaymentStrictSend != nil {
		if h.isRegulatedAsset(middleOp.PathPaymentStrictSend.SendAsset) {
			sendAmountFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictSend.SendAmount, 64)
			if err != nil {
				return 0, err
			}
			sendAmountInt64 := int64(math.Ceil(sendAmountFloat64))
			return sendAmountInt64, nil
		} else {
			destMinFloat64, err := strconv.ParseFloat(middleOp.PathPaymentStrictSend.DestMin, 64)
			if err != nil {
				return 0, err
			}
			destMinInt64 := int64(math.Ceil(destMinFloat64))
			return destMinInt64, nil
		}
	} else if middleOp.ManageSellOffer != nil {
		sellAmountFloat64, err := strconv.ParseFloat(middleOp.ManageSellOffer.Amount, 64)
		if err != nil {
			return 0, err
		}
		if h.isRegulatedAsset(middleOp.ManageSellOffer.Selling) {
			return int64(math.Ceil(sellAmountFloat64)), nil
		} else {
			priceFloat64 := float64(middleOp.ManageSellOffer.Price.N) / float64(middleOp.ManageSellOffer.Price.D)
			buyAmountFloat64 := math.Ceil(sellAmountFloat64 * priceFloat64)
			buyAmountInt64 := int64(buyAmountFloat64)
			return buyAmountInt64, nil
		}
	} else if middleOp.ManageBuyOffer != nil {
		buyAmountFloat64, err := strconv.ParseFloat(middleOp.ManageBuyOffer.Amount, 64)
		if err != nil {
			return 0, err
		}
		if h.isRegulatedAsset(middleOp.ManageBuyOffer.Buying) {
			return int64(math.Ceil(buyAmountFloat64)), nil
		} else {
			priceFloat64 := float64(middleOp.ManageBuyOffer.Price.N) / float64(middleOp.ManageBuyOffer.Price.D)
			sellAmountFloat64 := math.Ceil(buyAmountFloat64 * priceFloat64)
			sellAmountInt64 := int64(sellAmountFloat64)
			return sellAmountInt64, nil
		}
	} else {
		return 0, errors.New("middleOp has no operation set")
	}
}
