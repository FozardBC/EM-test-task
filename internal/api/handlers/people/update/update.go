package update

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"test-task/internal/domain/models"
	"test-task/internal/lib/api/response"
	"test-task/internal/storage"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Surname     *string `json:"surname,omitempty" validate:"omitempty,min=2,max=50"`
	Patronymic  *string `json:"patronymic,omitempty" validate:"omitempty,min=2,max=50"`
	Age         *int    `json:"age,omitempty"`
	Gender      *string `json:"gender,omitempty"`
	Nationality *string `json:"nationality,omitempty"`
}

type Response struct {
	Respone response.Response
	Person  models.Person
}

type PersonUpdater interface {
	Update(ctx context.Context, entity *models.Person, id int64) error
}

type UserProvider interface {
	FindByID(ctx context.Context, id int64) (*models.Person, error)
}

// Update
//
// @Summary 	Update person data
// @Description Update any field of person
// @Description Need at least one field to update
// @Tags 		people
// @ID 			update
// @Accept 		json
// @Produce 	json
// @Param		input		body		Request true "Person field data to update"
// @Param       id			query		string	true  "id of person to update"
// @Success 200 {object}	Response	"OK - Returns a person fields with update"
// @Failure 	400 "Invalid input"
// @Failure 	500 "Internal error"
// @Router 		/people [patch]
func New(log *slog.Logger, Provider UserProvider, Updater PersonUpdater) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := c.Request.Context()

		logHandler := log.With(
			slog.String("requestID", requestid.Get(c)),
		)

		var req Request

		prarmId := c.Param("id")

		id, err := strconv.Atoi(prarmId)
		if err != nil {
			logHandler.Error("can't param ID make int64")

			c.JSON(http.StatusBadRequest, response.Error("Invalid ID"))
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logHandler.Error("can't decode request body", "err", err.Error())

			c.JSON(http.StatusBadRequest, response.Error("failed to decode request body"))

			return
		}

		//валидация на наличие хотя бы одного поле
		err = req.Validate()
		if err != nil {
			logHandler.Error(err.Error())

			c.JSON(http.StatusBadRequest, response.Error(err.Error()))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validatorErr := err.(validator.ValidationErrors)

			logHandler.Error("invalid request", "err", err.Error())

			c.JSON(http.StatusBadRequest, response.ValidationError(validatorErr))

			return
		}

		person, err := Provider.FindByID(ctx, int64(id))
		if err != nil {
			if errors.Is(err, storage.ErrIDNotFound) {
				logHandler.Error("personID not found", "err", err.Error())

				c.JSON(http.StatusNoContent, nil)

				return
			}
			logHandler.Error("can't find person", "err", err.Error())

			c.JSON(http.StatusInternalServerError, response.Error("Internal server error")) //TODO: УЗНАТЬ ЧТО ЛУЧШЕ ВОЗВРАЩАТЬ В ТЕКСТЕ ОШИБКИ. 500 или описание ошибки

			return
		}

		checkForUpdates(logHandler, req, person)

		err = Updater.Update(ctx, person, int64(id))
		if err != nil {
			if errors.Is(err, storage.ErrIDNotFound) {
				logHandler.Error("personID not found", "err", err.Error())

				c.JSON(http.StatusNoContent, nil)

				return
			}
			logHandler.Error("can't find person", "err", err.Error())

			c.JSON(http.StatusInternalServerError, response.Error("Internal server error")) //TODO: УЗНАТЬ ЧТО ЛУЧШЕ ВОЗВРАЩАТЬ В ТЕКСТЕ ОШИБКИ. 500 или описание ошибки

			return
		}

		logHandler.Info("person updated", "Person", person, "id", id)

		c.JSON(http.StatusOK, Response{Respone: response.OK(), Person: *person})

	}
}

// 0_0
func checkForUpdates(log *slog.Logger, req Request, person *models.Person) {

	if req.Name != nil {
		person.Name = *req.Name
		log.Debug("Field Name changed")
	}
	if req.Surname != nil {
		person.Surname = *req.Surname
		log.Debug("Field Surname changed")
	}
	if req.Patronymic != nil {
		person.Patronymic = *req.Patronymic
		log.Debug("Field Patronymic changed")
	}
	if req.Age != nil {
		person.Age = *req.Age
		log.Debug("Field Age changed")
	}
	if req.Gender != nil {
		person.Gender = *req.Gender
		log.Debug("Field Gender changed")
	}
	if req.Nationality != nil {
		person.Nationality = *req.Nationality
		log.Debug("Field Nationality changed")
	}

}

func (r Request) Validate() error {
	if r.Name == nil && r.Surname == nil && r.Patronymic == nil &&
		r.Age == nil && r.Gender == nil && r.Nationality == nil {
		return errors.New("at least one field must be provided")
	}
	return nil
}
