package fx_rates

import (
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/stellar/go/services/regulated-assets-approval-server/internal/serve/httperror"
	"github.com/stellar/go/services/regulated-assets-approval-server/internal/serve/utils"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/support/http/httpdecode"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/support/render/httpjson"
	"github.com/stellar/go/txnbuild"
)

type PostHandler struct {
	Db *sqlx.DB
}

type postHandlerRequest struct {
	AssetCode   string `json:"asset_code" db:"asset_code"`
	AssetIssuer string `json:"asset_issuer" db:"asset_issuer"`
	UsdRate     string `json:"usd_rate" db:"usd_rate"`
}

func (h PostHandler) validate() error {
	if h.Db == nil {
		return errors.New("database cannot be nil")
	}
	return nil
}

func (r postHandlerRequest) validate() error {
	var asset txnbuild.BasicAsset
	if r.AssetIssuer == "" {
		asset = txnbuild.NativeAsset{}
	} else {
		asset = txnbuild.CreditAsset{Code: r.AssetCode, Issuer: r.AssetIssuer}
	}
	err := utils.ValidateStellarAsset(asset)
	if err != nil {
		return err
	}
	_, err = strconv.ParseFloat(r.UsdRate, 64)
	return err
}

func (h PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.validate()
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "validating fx-rates PostHandler"))
		httperror.InternalServer.Render(w)
		return
	}

	in := postHandlerRequest{}
	err = httpdecode.Decode(r, &in)
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "decoding fx-rates POST Request"))
		httperror.BadRequest.Render(w)
		return
	}
	err = in.validate()
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "request fields contain an invalid value"))
		httperror.BadRequest.Render(w)
		return
	}
	const query = `
		WITH new_row AS (
			INSERT INTO fx_rates (asset_code, asset_issuer, usd_rate)
			VALUES ($1, $2, $3)
			RETURNING *
		)
		SELECT id, asset_code, asset_issuer, usd_rate, rate_timestamp FROM new_row
	`
	insertedFxRate := FxRate{}
	err = h.Db.QueryRowContext(
		ctx,
		query,
		in.AssetCode,
		in.AssetIssuer,
		in.UsdRate,
	).Scan(
		&insertedFxRate.Id,
		&insertedFxRate.AssetCode,
		&insertedFxRate.AssetIssuer,
		&insertedFxRate.UsdRate,
		&insertedFxRate.RateTimestamp,
	)
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "inserting new row into fx_rates table"))
		httperror.BadRequest.Render(w)
		return
	}
	httpjson.RenderStatus(w, 200, insertedFxRate, httpjson.JSON)
}
