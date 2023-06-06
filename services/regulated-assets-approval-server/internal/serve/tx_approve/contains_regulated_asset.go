package tx_approve

import (
	"errors"
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) getRegulatedAsset(
	middleOp *MiddleOperation,
) (txnbuild.Asset, error) {
	issuerAddress := h.IssuerKP.Address()
	if middleOp.ManageSellOffer != nil {
		offer := middleOp.ManageSellOffer
		if offer.Selling.GetCode() == h.AssetCode &&
			offer.Selling.GetIssuer() == issuerAddress {
			return offer.Selling, nil
		}
		if offer.Buying.GetCode() == h.AssetCode &&
			offer.Buying.GetIssuer() == issuerAddress {
			return offer.Buying, nil
		}
	} else if middleOp.ManageBuyOffer != nil {
		offer := middleOp.ManageBuyOffer
		if offer.Selling.GetCode() == h.AssetCode &&
			offer.Selling.GetIssuer() == issuerAddress {
			return offer.Selling, nil
		}
		if offer.Buying.GetCode() == h.AssetCode &&
			offer.Buying.GetIssuer() == issuerAddress {
			return offer.Buying, nil
		}
	} else if middleOp.Payment != nil {
		payment := middleOp.Payment
		if payment.Asset.GetCode() == h.AssetCode &&
			payment.Asset.GetIssuer() == issuerAddress {
			return payment.Asset, nil
		}
	}
	return nil, errors.New("failed to find asset")
}
