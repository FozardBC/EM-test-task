package create

import (
	"context"
	"log/slog"
	"net/http"
	"test-task/internal/domain/models"
	"test-task/internal/lib/api/response"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Name string `json:"name" validate:"required,min=2,max=50" example:"Alexander"`
	// required
	Surname string `json:"surname" validate:"required,min=2,max=50" example:"Sidorov"`
	// required
	Patronymic string `json:"patronymic,omitempty" validate:"omitempty,min=2,max=50" example:"Petrovich"`
	// omitempty
}

type Response struct {
	Resp response.Response `json:"response"`
	ID   int64             `json:"id,omitempty"`
}

type PersonSaver interface {
	Save(ctx context.Context, entity *models.Person) (int64, error)
}

type IEnricher interface {
	Enrich(ctx context.Context, person *models.Person) (*models.Person, error)
}

// Create godoc
//
// @Summary 	Create new user
// @Description Creating, enriching and saving new user
// @Tags 		people
// @ID 			create
// @Accept 		json
// @Produce 	json
// @Param		input body Request true "Person basic info"
// @Success 200 {object} Response "OK"
// @Failure 	400 "Invalid input"
// @Failure 	500 "Internal error"
// @Router 		/people [post]
func New(log *slog.Logger, Enricher IEnricher, Saver PersonSaver) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := c.Request.Context()

		logHandler := log.With(
			slog.String("requestID", requestid.Get(c)),
		)

		var req Request

		if err := c.BindJSON(&req); err != nil {
			logHandler.Error("can't decode request body", "err", err.Error())

			c.JSON(http.StatusBadRequest, response.Error("failed to decode request body"))

			return
		}

		if err := validator.New().Struct(req); err != nil {
			validatorErr := err.(validator.ValidationErrors)

			logHandler.Error("invalid request", "err", err.Error())

			c.JSON(http.StatusBadRequest, response.ValidationError(validatorErr))

			return
		}

		person := &models.Person{
			Name:       req.Name,
			Surname:    req.Surname,
			Patronymic: req.Patronymic,
		}

		person, err := Enricher.Enrich(ctx, person)
		if err != nil {
			logHandler.Error("can't enrich person", "err", err.Error())

			c.JSON(http.StatusInternalServerError, response.Error("Internal server error"))

			return
		}

		if person.Patronymic == "" || person.Patronymic == " " {
			person.Patronymic = "N/A"
		}

		id, err := Saver.Save(ctx, person)
		if err != nil {
			logHandler.Error("can't save person", "err", err.Error())

			c.JSON(http.StatusInternalServerError, response.Error("Internal server error")) //TODO: УЗНАТЬ ЧТО ЛУЧШЕ ВОЗВРАЩАТЬ В ТЕКСТЕ ОШИБКИ. 500 или описание ошибки

			return
		}

		logHandler.Info("person saved", "Person", person, "id", id)

		c.JSON(http.StatusOK, Response{Resp: response.OK(), ID: id})

	}
}
