package tx_approve

import (
	"github.com/stellar/go/txnbuild"
)

func containsTrustLineFlags(
	flags []txnbuild.TrustLineFlag,
	requiredFlags []txnbuild.TrustLineFlag,
) bool {
	for _, requiredFlag := range requiredFlags {
		contains := false
		for _, flag := range flags {
			if requiredFlag == flag {
				contains = true
				break
			}
		}
		if !contains {
			return false
		}
	}
	return true
}
