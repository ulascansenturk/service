package v1

type API struct {
	transfersService *TransfersService
	usersService     *UsersService
}

func NewAPI(transfersService *TransfersService, usersService *UsersService) *API {
	return &API{
		transfersService: transfersService,
		usersService:     usersService,
	}
}
