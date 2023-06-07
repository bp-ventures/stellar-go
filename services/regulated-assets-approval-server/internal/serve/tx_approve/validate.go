package tx_approve

import (
	"github.com/stellar/go/support/errors"
)

// validate performs some validations on the provided handler data.
func (h TxApprove) validate() error {
	if h.IssuerKP == nil {
		return errors.New("issuer keypair cannot be nil")
	}
	if h.AssetCode == "" {
		return errors.New("asset code cannot be empty")
	}
	if h.HorizonClient == nil {
		return errors.New("horizon client cannot be nil")
	}
	if h.NetworkPassphrase == "" {
		return errors.New("network passphrase cannot be empty")
	}
	if h.Db == nil {
		return errors.New("database cannot be nil")
	}
	if h.KycPaymentThreshold <= 0 {
		return errors.New("kyc threshold cannot be less than or equal to zero")
	}
	if h.BaseURL == "" {
		return errors.New("base url cannot be empty")
	}
	return nil
}
