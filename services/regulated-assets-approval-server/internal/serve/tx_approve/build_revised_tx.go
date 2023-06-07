package tx_approve

import (
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) buildRevisedTx(
	acc txnbuild.Account,
	middleOp *MiddleOperation,
) (string, error) {
	asset, _ := h.getRegulatedAsset(middleOp)
	issuerAddress := h.IssuerKP.Address()
	var revisedOperations []txnbuild.Operation
	if middleOp.Payment != nil {
		revisedOperations = h.buildRevisedOperationsPayment(
			acc,
			middleOp,
			asset,
			issuerAddress,
		)
	} else if middleOp.PathPaymentStrictReceive != nil {
		revisedOperations = h.buildRevisedOperationsPathReceive(
			acc,
			middleOp,
			asset,
			issuerAddress,
		)
	} else if middleOp.PathPaymentStrictSend != nil {
		revisedOperations = h.buildRevisedOperationsPathSend(
			acc,
			middleOp,
			asset,
			issuerAddress,
		)
	} else if middleOp.ManageSellOffer != nil {
		revisedOperations = h.buildRevisedOperationsManageSellOffer(
			acc,
			middleOp,
			asset,
			issuerAddress,
		)
	} else if middleOp.ManageBuyOffer != nil {
		revisedOperations = h.buildRevisedOperationsManageBuyOffer(
			acc,
			middleOp,
			asset,
			issuerAddress,
		)
	} else {
		return "", errors.New("middle operation has all nil")
	}
	revisedTx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        acc,
		IncrementSequenceNum: true,
		Operations:           revisedOperations,
		BaseFee:              300,
		Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
	})
	revisedTx, err = revisedTx.Sign(h.NetworkPassphrase, h.IssuerKP)
	if err != nil {
		return "", errors.Wrap(err, "signing transaction")
	}
	revisedTxXdr, err := revisedTx.Base64()
	if err != nil {
		return "", errors.Wrap(err, "encoding revised transaction")
	}
	return revisedTxXdr, nil
}
