package fx_rates

type FxRate struct {
	Id            string `json:"id" db:"id"`
	AssetCode     string `json:"asset_code" db:"asset_code"`
	AssetIssuer   string `json:"asset_issuer" db:"asset_issuer"`
	UsdRate       string `json:"usd_rate" db:"usd_rate"`
	RateTimestamp string `json:"rate_timestamp" db:"rate_timestamp"`
}
