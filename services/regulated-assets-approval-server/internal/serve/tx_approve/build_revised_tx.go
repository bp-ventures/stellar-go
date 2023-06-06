package tx_approve

import (
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/txnbuild"
)

func (h TxApprove) buildRevisedTx(
	acc txnbuild.Account,
	middleOp *MiddleOperation,
) (string, error) {
	asset, err := h.getRegulatedAsset(middleOp)
	if err != nil {
		return "", errors.New("asset is not supported")
	}
	issuerAddress := h.IssuerKP.Address()
	var revisedOperations []txnbuild.Operation
	if middleOp.ManageSellOffer != nil {
		revisedOperations = []txnbuild.Operation{
			&txnbuild.SetTrustLineFlags{
				SourceAccount: issuerAddress,
				Trustor:       middleOp.SourceAccount,
				Asset:         asset,
				SetFlags: []txnbuild.TrustLineFlag{
					txnbuild.TrustLineAuthorized,
					txnbuild.TrustLineAuthorizedToMaintainLiabilities,
				},
			},
			middleOp.ManageSellOffer,
			&txnbuild.SetTrustLineFlags{
				SourceAccount: issuerAddress,
				Trustor:       middleOp.SourceAccount,
				Asset:         asset,
				ClearFlags: []txnbuild.TrustLineFlag{
					txnbuild.TrustLineAuthorized,
					// Here we'd add txnbuild.TrustLineAuthorizedToMaintainLiabilities,
					// but this would prevent the account from creating offers.
				},
			},
		}
	} else if middleOp.ManageBuyOffer != nil {
		revisedOperations = []txnbuild.Operation{
			&txnbuild.SetTrustLineFlags{
				SourceAccount: issuerAddress,
				Trustor:       middleOp.SourceAccount,
				Asset:         asset,
				SetFlags: []txnbuild.TrustLineFlag{
					txnbuild.TrustLineAuthorized,
					txnbuild.TrustLineAuthorizedToMaintainLiabilities,
				},
			},
			middleOp.ManageBuyOffer,
			&txnbuild.SetTrustLineFlags{
				SourceAccount: issuerAddress,
				Trustor:       middleOp.SourceAccount,
				Asset:         asset,
				ClearFlags: []txnbuild.TrustLineFlag{
					txnbuild.TrustLineAuthorized,
					// Here we'd add txnbuild.TrustLineAuthorizedToMaintainLiabilities,
					// but this would prevent the account from creating offers.
				},
			},
		}
	} else if middleOp.Payment != nil {
		revisedOperations = []txnbuild.Operation{
			&txnbuild.SetTrustLineFlags{
				SourceAccount: issuerAddress,
				Trustor:       middleOp.SourceAccount,
				Asset:         asset,
				SetFlags: []txnbuild.TrustLineFlag{
					txnbuild.TrustLineAuthorized,
					txnbuild.TrustLineAuthorizedToMaintainLiabilities,
				},
			},
			&txnbuild.SetTrustLineFlags{
				SourceAccount: issuerAddress,
				Trustor:       middleOp.Payment.Destination,
				Asset:         asset,
				SetFlags: []txnbuild.TrustLineFlag{
					txnbuild.TrustLineAuthorized,
					txnbuild.TrustLineAuthorizedToMaintainLiabilities,
				},
			},
			middleOp.Payment,
			&txnbuild.SetTrustLineFlags{
				SourceAccount: issuerAddress,
				Trustor:       middleOp.SourceAccount,
				Asset:         asset,
				ClearFlags: []txnbuild.TrustLineFlag{
					txnbuild.TrustLineAuthorized,
					// Here we'd add txnbuild.TrustLineAuthorizedToMaintainLiabilities,
					// but this would prevent the account from creating offers.
				},
			},
			&txnbuild.SetTrustLineFlags{
				SourceAccount: issuerAddress,
				Trustor:       middleOp.Payment.Destination,
				Asset:         asset,
				ClearFlags: []txnbuild.TrustLineFlag{
					txnbuild.TrustLineAuthorized,
					// Here we'd add txnbuild.TrustLineAuthorizedToMaintainLiabilities,
					// but this would prevent the account from creating offers.
				},
			},
		}
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
