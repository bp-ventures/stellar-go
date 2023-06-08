package tx_approve

import (
	"errors"

	"github.com/shopspring/decimal"
)

func (h TxApprove) parseAmount(middleOp *MiddleOperation) (int64, error) {
	if middleOp.Payment != nil {
		if amount, err := decimal.NewFromString(middleOp.Payment.Amount); err != nil {
			return 0, err
		} else {
			return amount.Ceil().IntPart(), nil
		}
	} else if middleOp.PathPaymentStrictReceive != nil {
		if h.isRegulatedAsset(middleOp.PathPaymentStrictReceive.SendAsset) {
			sendMax, err := decimal.NewFromString(middleOp.PathPaymentStrictReceive.SendMax)
			if err != nil {
				return 0, err
			}
			return sendMax.Ceil().IntPart(), nil
		} else {
			destAmount, err := decimal.NewFromString(middleOp.PathPaymentStrictReceive.DestAmount)
			if err != nil {
				return 0, err
			}
			return destAmount.Ceil().IntPart(), nil
		}
	} else if middleOp.PathPaymentStrictSend != nil {
		if h.isRegulatedAsset(middleOp.PathPaymentStrictSend.SendAsset) {
			sendAmount, err := decimal.NewFromString(middleOp.PathPaymentStrictSend.SendAmount)
			if err != nil {
				return 0, err
			}
			return sendAmount.Ceil().IntPart(), nil
		} else {
			destMin, err := decimal.NewFromString(middleOp.PathPaymentStrictSend.DestMin)
			if err != nil {
				return 0, err
			}
			return destMin.Ceil().IntPart(), nil
		}
	} else if middleOp.ManageSellOffer != nil {
		sellAmount, err := decimal.NewFromString(middleOp.ManageSellOffer.Amount)
		if err != nil {
			return 0, err
		}
		if h.isRegulatedAsset(middleOp.ManageSellOffer.Selling) {
			return sellAmount.Ceil().IntPart(), nil
		} else {
			numerator := decimal.NewFromInt32(int32(middleOp.ManageSellOffer.Price.N))
			denominator := decimal.NewFromInt32(int32(middleOp.ManageSellOffer.Price.D))
			price := numerator.Div(denominator)
			buyAmount := sellAmount.Mul(price)
			return buyAmount.Ceil().IntPart(), nil
		}
	} else if middleOp.ManageBuyOffer != nil {
		buyAmount, err := decimal.NewFromString(middleOp.ManageBuyOffer.Amount)
		if err != nil {
			return 0, err
		}
		if h.isRegulatedAsset(middleOp.ManageBuyOffer.Buying) {
			return buyAmount.Ceil().IntPart(), nil
		} else {
			numerator := decimal.NewFromInt32(int32(middleOp.ManageBuyOffer.Price.N))
			denominator := decimal.NewFromInt32(int32(middleOp.ManageBuyOffer.Price.D))
			price := numerator.Div(denominator)
			sellAmount := buyAmount.Mul(price)
			return sellAmount.Ceil().IntPart(), nil
		}
	} else {
		return 0, errors.New("middleOp has no operation set")
	}
}
