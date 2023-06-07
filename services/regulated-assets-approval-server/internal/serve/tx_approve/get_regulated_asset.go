package tx_approve

import (
	"errors"
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) getRegulatedAsset(
	middleOp *MiddleOperation,
) (txnbuild.Asset, error) {
	if middleOp.PathPaymentStrictReceive != nil {
		pathPaymentStrictReceive := middleOp.PathPaymentStrictReceive
		if h.isRegulatedAsset(pathPaymentStrictReceive.SendAsset) {
			return pathPaymentStrictReceive.SendAsset, nil
		}
		if h.isRegulatedAsset(pathPaymentStrictReceive.DestAsset) {
			return pathPaymentStrictReceive.DestAsset, nil
		}
	} else if middleOp.PathPaymentStrictSend != nil {
		pathPaymentStrictSend := middleOp.PathPaymentStrictSend
		if h.isRegulatedAsset(pathPaymentStrictSend.SendAsset) {
			return pathPaymentStrictSend.SendAsset, nil
		}
		if h.isRegulatedAsset(pathPaymentStrictSend.DestAsset) {
			return pathPaymentStrictSend.DestAsset, nil
		}
	} else if middleOp.Payment != nil {
		payment := middleOp.Payment
		if h.isRegulatedAsset(payment.Asset) {
			return payment.Asset, nil
		}
	} else if middleOp.ManageSellOffer != nil {
		offer := middleOp.ManageSellOffer
		if h.isRegulatedAsset(offer.Selling) {
			return offer.Selling, nil
		}
		if h.isRegulatedAsset(offer.Buying) {
			return offer.Buying, nil
		}
	} else if middleOp.ManageBuyOffer != nil {
		offer := middleOp.ManageBuyOffer
		if h.isRegulatedAsset(offer.Selling) {
			return offer.Selling, nil
		}
		if h.isRegulatedAsset(offer.Buying) {
			return offer.Buying, nil
		}
	}
	return nil, errors.New("failed to find asset")
}
