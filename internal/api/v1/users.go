package v1

import (
	"context"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"net/http"
	"ulascansenturk/service/internal/accounts"
	"ulascansenturk/service/internal/api/server"
	"ulascansenturk/service/internal/constants"
	"ulascansenturk/service/internal/users"
)

type UsersService struct {
	service        users.Service
	accountService accounts.Service
}

func NewUsersService(service users.Service, accountsSErvice accounts.Service) *UsersService {
	return &UsersService{service: service, accountService: accountsSErvice}
}

func (a *API) V1CreateUser(w http.ResponseWriter, r *http.Request) {
	reqBody := new(server.V1CreateUserJSONRequestBody)

	err := render.Bind(r, reqBody)
	if err != nil {
		server.BadRequestError(err, w, r)

		return
	}

	result, err := a.usersService.createUserWithBankAccount(r.Context(), reqBody)

	if err != nil {
		log.Err(err).Msg("user processing failed")

		server.ProcessingError(err, w, r)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, result)

}

func (a *UsersService) createUserWithBankAccount(ctx context.Context, reqBody *server.V1CreateUserJSONRequestBody) (*server.UserResult, error) {
	tx := GetDBFromContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	user, userCreateErr := a.service.CreateUser(ctx, &users.User{
		ID:        uuid.New(),
		Email:     reqBody.Data.Email,
		FirstName: reqBody.Data.FirstName,
		LastName:  reqBody.Data.LastName,
	}, reqBody.Data.Password)
	if userCreateErr != nil {
		tx.Rollback()
		return nil, userCreateErr
	}

	bankAccount, bankAccountErr := a.accountService.CreateAccount(ctx, &accounts.Account{
		ID:       uuid.New(),
		UserID:   user.ID,
		Balance:  *reqBody.Data.Balance,
		Status:   constants.AccountStatusACTIVE,
		Currency: reqBody.Data.CurrencyCode,
	})
	if bankAccountErr != nil {
		tx.Rollback()
		return nil, bankAccountErr
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &server.UserResult{
		BankAccount: &server.Account{
			Balance:  int32(bankAccount.Balance),
			Currency: bankAccount.Currency,
			Id:       &bankAccount.ID,
			Status:   bankAccount.Status.String(),
			UserId:   user.ID,
		},
		User: &server.User{
			Email:     types.Email(user.Email),
			FirstName: user.FirstName,
			Id:        &user.ID,
			LastName:  user.LastName,
		},
	}, nil
}

func GetDBFromContext(ctx context.Context) *gorm.DB {
	if db, ok := ctx.Value("gormDB").(*gorm.DB); ok {
		return db
	}
	return nil
}
