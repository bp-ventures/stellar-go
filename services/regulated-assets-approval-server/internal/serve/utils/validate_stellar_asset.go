package utils

import (
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/txnbuild"
)

// validateStellarPublicKey returns an error if a public key is invalid. Otherwise, it returns nil.
// It is a wrapper around the IsValidEd25519PublicKey method of the strkey package.
func ValidateStellarPublicKey(publicKey string) error {
	if publicKey == "" {
		return errors.New("public key is undefined")
	}

	if !strkey.IsValidEd25519PublicKey(publicKey) {
		return errors.Errorf("%s is not a valid stellar public key", publicKey)
	}
	return nil
}

// validateStellarAsset checks if the asset supplied is a valid stellar Asset. It returns an error if the asset is
// nil, has an invalid asset code or issuer.
func ValidateStellarAsset(asset txnbuild.BasicAsset) error {
	if asset == nil {
		return errors.New("asset is undefined")
	}

	if asset.IsNative() {
		return nil
	}

	_, err := asset.GetType()
	if err != nil {
		return err
	}

	err = ValidateStellarPublicKey(asset.GetIssuer())
	if err != nil {
		return errors.Errorf("asset issuer: %s", err.Error())
	}

	return nil
}
