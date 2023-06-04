-- +migrate Up

CREATE TABLE public.fx_rates (
    id SERIAL PRIMARY KEY,
    base_asset_code text NOT NULL,
    base_asset_issuer text NOT NULL,
    quote_asset_code text NOT NULL,
    quote_asset_issuer text NOT NULL,
    rate text NOT NULL,  -- amount of quote asset that buys 1 unit of base asset
    rate_timestamp timestamp with time zone NOT NULL DEFAULT NOW(),
);

-- +migrate Down

DROP TABLE public.fx_rates;

