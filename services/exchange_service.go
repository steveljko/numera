package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"numera/model"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

var CacheTTL = 1 * time.Hour

type Response struct {
	Data struct {
		Mid  float64 `json:"mid"`
		Date string  `json:"date"`
	} `json:"data"`
}

type cacheEntry struct {
	rate      decimal.Decimal
	expiresAt time.Time
}

type ExchangeService struct {
	baseURL    string
	httpClient *http.Client
	cache      sync.Map
	logger     *logrus.Logger
}

func NewExchangeService(logger *logrus.Logger) *ExchangeService {
	return &ExchangeService{
		baseURL:    "https://hexarate.paikama.co/api/rates",
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

func (es *ExchangeService) getCacheKey(from, to model.Currency) string {
	return fmt.Sprintf("%s:%s", from, to)
}

// fetchRate retrieves the current exchange rate from api for the given currency pair.
func (es *ExchangeService) fetchRate(
	ctx context.Context,
	from, to model.Currency,
) (decimal.Decimal, error) {
	if from == "" || to == "" {
		es.logger.WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).Error("empty_currency_codes")
		return decimal.Zero, errors.New("currency codes must not be empty")
	}

	cacheKey := es.getCacheKey(from, to)

	if cached, ok := es.cache.Load(cacheKey); ok {
		entry := cached.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			es.logger.WithFields(logrus.Fields{
				"from":       from,
				"to":         to,
				"rate":       entry.rate,
				"expires_at": entry.expiresAt,
			}).Debug("exchange_rate_cache_hit")
			return entry.rate, nil
		}

		es.cache.Delete(cacheKey)
		es.logger.WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).Debug("exchange_rate_cache_expired")
	} else {
		es.logger.WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).Debug("exchange_rate_cache_miss")
	}

	url := fmt.Sprintf("%s/%s/%s/latest", es.baseURL, from, to)

	es.logger.WithFields(logrus.Fields{
		"url":  url,
		"from": from,
		"to":   to,
	}).Debug("making_http_request_to_exchange_api")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		es.logger.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("failed_to_create_http_request")
		return decimal.Zero, err
	}

	resp, err := es.httpClient.Do(req)
	if err != nil {
		es.logger.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("http_request_failed")
		return decimal.Zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		es.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"body":        string(body),
			"url":         url,
		}).Error("exchange_api_returned_non_ok_status")
		return decimal.Zero, fmt.Errorf(
			"exchange api error: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		es.logger.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("failed_to_decode_api_response")
		return decimal.Zero, err
	}

	rate := decimal.NewFromFloat(result.Data.Mid)

	expiresAt := time.Now().Add(CacheTTL)
	es.cache.Store(cacheKey, cacheEntry{
		rate:      rate,
		expiresAt: expiresAt,
	})

	es.logger.WithFields(logrus.Fields{
		"from":       from,
		"to":         to,
		"rate":       rate,
		"date":       result.Data.Date,
		"expires_at": expiresAt,
		"cache_ttl":  CacheTTL,
	}).Info("exchange_rate_fetched_and_cached")

	return rate, nil
}

// ConvertAmount converts a amount from one currency to another using the exchange rate.
func (es *ExchangeService) ConvertAmount(
	ctx context.Context,
	amount decimal.Decimal,
	from, to model.Currency,
) (decimal.Decimal, error) {
	es.logger.WithFields(logrus.Fields{
		"amount": amount,
		"from":   from,
		"to":     to,
	}).Debug("converting_amount")

	rate, err := es.fetchRate(ctx, from, to)
	if err != nil {
		es.logger.WithFields(logrus.Fields{
			"amount": amount,
			"from":   from,
			"to":     to,
		}).WithError(err).Error("failed_to_get_exchange_rate_for_conversion")
		return decimal.Zero, err
	}

	converted := amount.Mul(rate).Round(2)

	es.logger.WithFields(logrus.Fields{
		"amount":    amount,
		"from":      from,
		"to":        to,
		"rate":      rate,
		"converted": converted,
	}).Info("amount_converted_successfully")

	return converted, nil
}

// ClearCache clears all cached rates
func (es *ExchangeService) ClearCache() {
	es.logger.Info("clearing_all_cached_exchange_rates")
	es.cache.Range(func(key, value any) bool {
		es.cache.Delete(key)
		return true
	})
	es.logger.Info("cache_cleared_successfully")
}

// ClearCacheForPair clears cached rate for a specific currency pair
func (es *ExchangeService) ClearCacheForPair(from, to model.Currency) {
	cacheKey := es.getCacheKey(from, to)
	es.cache.Delete(cacheKey)
	es.logger.WithFields(logrus.Fields{
		"from": from,
		"to":   to,
	}).Info("cache_cleared_for_currency_pair")
}
