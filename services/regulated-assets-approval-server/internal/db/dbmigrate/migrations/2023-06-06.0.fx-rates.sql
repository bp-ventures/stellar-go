-- +migrate Up

CREATE TABLE public.fx_rates (
    id SERIAL PRIMARY KEY,
    asset_code text NOT NULL,
    asset_issuer text NOT NULL,
    usd_rate text NOT NULL,  -- amount of asset that corresponds to 1 US Dollar
    rate_timestamp timestamp with time zone NOT NULL DEFAULT NOW(),
);

-- +migrate Down

DROP TABLE public.fx_rates;

