package fx_rates

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/stellar/go/services/regulated-assets-approval-server/internal/serve/httperror"
	"github.com/stellar/go/support/render/httpjson"
)

type GetHandler struct {
	Db *sqlx.DB
}

type GetHandlerResponse struct {
	Data []FxRate `json:"data"`
}

type FxRate struct {
	Id            string `json:"id" db:"id"`
	AssetCode     string `json:"asset_code" db:"asset_code"`
	AssetIssuer   string `json:"asset_issuer" db:"asset_issuer"`
	UsdRate       string `json:"usd_rate" db:"usd_rate"`
	RateTimestamp string `json:"rate_timestamp" db:"rate_timestamp"`
}

func (h GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const query = `
		SELECT id, asset_code, asset_issuer, usd_rate, rate_timestamp
		FROM fx_rates
		ORDER BY rate_timestamp DESC
	`
	var fxRates []FxRate
	if err := h.Db.SelectContext(ctx, &fxRates, query); err != nil {
		httperror.InternalServer.Render(w)
		return
	}
	resp := GetHandlerResponse{Data: fxRates}
	httpjson.RenderStatus(w, 200, resp, httpjson.JSON)
}
