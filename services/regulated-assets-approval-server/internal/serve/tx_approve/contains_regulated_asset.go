package tx_approve

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
