package fx_rates

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/stellar/go/services/regulated-assets-approval-server/internal/serve/httperror"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/support/render/httpjson"
)

type GetHandler struct {
	Db *sqlx.DB
}

type GetHandlerResponse struct {
	Data []FxRate `json:"data"`
}

func (h GetHandler) validate() error {
	if h.Db == nil {
		return errors.New("database cannot be nil")
	}
	return nil
}

func (h GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.validate()
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "validating fx-rates GetHandler"))
		httperror.InternalServer.Render(w)
		return
	}
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
