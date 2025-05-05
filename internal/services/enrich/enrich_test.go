package enrich

import (
	"context"
	"log/slog"
	"net/http"
	"reflect"
	"test-task/internal/domain/models"
	"test-task/internal/logger"
	"testing"
)

func TestEnricher_Enrich(t *testing.T) {
	type fields struct {
		log    *slog.Logger
		client *http.Client
	}
	type args struct {
		ctx    context.Context
		person *models.Person
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Person
		wantErr bool
	}{
		{
			fields: fields{
				log:    logger.New("debug"),
				client: &http.Client{},
			},
			args: args{
				person: &models.Person{
					Name:       "Oleg",
					Surname:    "Petrov",
					Patronymic: "Ivanovich",
				},
				ctx: context.TODO(),
			},
			want: &models.Person{
				Name:        "Oleg",
				Surname:     "Petrov",
				Patronymic:  "Ivanovich",
				Age:         55,
				Gender:      "male",
				Nationality: "UA",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Enricher{
				log:    tt.fields.log,
				client: tt.fields.client,
			}
			got, err := e.Enrich(tt.args.ctx, tt.args.person)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enricher.Enrich() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Enricher.Enrich() = %v, want %v", got, tt.want)
			}
		})
	}
}
