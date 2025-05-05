package enrich

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"test-task/internal/domain/models"
	"time"
)

const (
	AgeAPI         = "https://api.agify.io/?name="
	GenderAPI      = "https://api.genderize.io/?name="
	NationalityAPI = "https://api.nationalize.io/?name="
	timeout        = 60 * time.Second
)

type Enricher struct {
	log    *slog.Logger
	client *http.Client
}

func New(log *slog.Logger) *Enricher {
	return &Enricher{
		client: &http.Client{Timeout: timeout},
		log:    log,
	}
}

func (e *Enricher) Enrich(ctx context.Context, person *models.Person) (*models.Person, error) {

	var wg sync.WaitGroup
	ageResult := make(chan int, 1)
	genderResult := make(chan string, 1)
	nationalityResult := make(chan string, 1)

	errChan := make(chan error, 3)

	wg.Add(3)
	go e.fetchAge(ctx, person.Name, &wg, ageResult, errChan)
	go e.fetchGender(ctx, person.Name, &wg, genderResult, errChan)
	go e.fetchNationality(ctx, person.Name, &wg, nationalityResult, errChan)

	wg.Wait()

	close(errChan)
	close(ageResult)
	close(genderResult)
	close(nationalityResult)

	err, opened := <-errChan
	if err != nil && !opened {
		return nil, fmt.Errorf("failed to enrich person data: %w", err)
	}

	person.Age = <-ageResult
	person.Gender = <-genderResult
	person.Nationality = <-nationalityResult

	return person, nil
}
