package enrich

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func (e *Enricher) fetchAge(ctx context.Context, name string, wg *sync.WaitGroup, resultChan chan<- int, errChan chan<- error) {

	defer wg.Done()
	url := AgeAPI + name
	var result struct {
		Age int `json:"age"`
	}

	e.log.Debug("request to fetch age", "name", name, "url", url)

	if err := e.fetchAPI(ctx, url, &result); err != nil {
		e.log.Error("failed to fetch age", "error", err, "name", name)
		errChan <- fmt.Errorf("age API: %w", err)
		return
	}

	e.log.Debug("fetched gender", "result", result.Age)

	resultChan <- result.Age
}

func (e *Enricher) fetchGender(ctx context.Context, name string, wg *sync.WaitGroup, resultChan chan<- string, errChan chan<- error) {
	defer wg.Done()
	url := GenderAPI + name
	var result struct {
		Gender string `json:"gender"`
	}

	e.log.Debug("request to fetch gender", "name", name, "url", url)

	if err := e.fetchAPI(ctx, url, &result); err != nil {
		e.log.Error("failed to fetch age", "error", err, "name", name)
		errChan <- fmt.Errorf("age API err: %w", err)
		return
	}

	e.log.Debug("fetched gender", "result", result.Gender)

	resultChan <- result.Gender
}

func (e *Enricher) fetchNationality(ctx context.Context, name string, wg *sync.WaitGroup, resultChan chan<- string, errChan chan<- error) {
	defer wg.Done()
	url := NationalityAPI + name
	type countryEntity struct {
		CountryID   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	}

	var result struct {
		Countries []countryEntity `json:"country"`
	}

	e.log.Debug("request to fetch nationality", "name", name, "url", url)

	if err := e.fetchAPI(ctx, url, &result); err != nil {
		e.log.Error("failed to fetch nationality", "error", err, "name", name)
		errChan <- fmt.Errorf("nationality API err: %w", err)
		return
	}

	if len(result.Countries) < 1 {
		err := errors.New("len of nationality less 1")

		e.log.Error("failed to fetch nationality", "error", err, "name", name)
		errChan <- fmt.Errorf("nationality API err: %w", err)
		return
	}

	e.log.Debug("fetched nationality", "result", result.Countries[0].CountryID)

	resultChan <- result.Countries[0].CountryID
}

func (e *Enricher) fetchAPI(ctx context.Context, url string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		e.log.Error("can't get request")

		return fmt.Errorf("can't get request:%w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		e.log.Error("can't get request")

		return fmt.Errorf("request failed: %w", err)

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e.log.Error("can't get request")

		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e.log.Error("can't get request")

		return fmt.Errorf("read body failed: %w", err)

	}

	if err := json.Unmarshal(body, target); err != nil {
		e.log.Error("can't get request")

		return fmt.Errorf("json unmarshal failed: %w", err)

	}

	return nil
}
