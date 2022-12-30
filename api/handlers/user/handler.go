package user

import (
	"bjungle-consenso/internal/env"
	"bjungle-consenso/internal/grpc/accounting_proto"
	"bjungle-consenso/internal/grpc/auth_proto"
	"bjungle-consenso/internal/grpc/users_proto"
	"bjungle-consenso/internal/grpc/wallet_proto"
	"bjungle-consenso/internal/helpers"
	"bjungle-consenso/internal/logger"
	"bjungle-consenso/internal/msg"
	"bjungle-consenso/pkg/bc"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcMetadata "google.golang.org/grpc/metadata"
	"net/http"
)

type handlerUser struct {
	DB   *sqlx.DB
	TxID string
}

func (h *handlerUser) Login(c *fiber.Ctx) error {
	res := responseLogin{Error: true}
	request := rqLogin{}
	e := env.NewConfiguration()
	err := c.BodyParser(&request)
	if err != nil {
		logger.Error.Printf("couldn't bind model rqLogin: %v", err)
		res.Code, res.Type, res.Msg = 22, 1, "La estructura enviada no es valida"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientAuth := auth_proto.NewAuthServicesUsersClient(connAuth)

	resLogin, err := clientAuth.Login(context.Background(), &auth_proto.LoginRequest{
		Email:    &request.Email,
		Nickname: &request.NickName,
		Password: request.Password,
	})
	if err != nil {
		logger.Error.Printf("No se pudo obtener el token de autenticacion: %v", err)
		res.Code, res.Type, res.Msg = 22, 1, "No se pudo obtener el token de autenticacion"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resLogin == nil {
		logger.Error.Printf("No se pudo obtener el token de autenticacion")
		res.Code, res.Type, res.Msg = 22, 1, "No se pudo obtener el token de autenticacion"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resLogin.Error {
		logger.Error.Printf(resLogin.Msg)
		res.Code, res.Type, res.Msg = msg.GetByCode(22, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	authRes := Token{AccessToken: resLogin.Data.AccessToken, RefreshToken: resLogin.Data.RefreshToken}
	res.Data = authRes
	res.Code, res.Type, res.Msg = msg.GetByCode(29, h.DB, h.TxID)
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

func (h *handlerUser) GetWalletsByUserId(c *fiber.Ctx) error {

	res := resGetWallets{Error: true}
	e := env.NewConfiguration()
	srvBk := bc.NewServerBk(h.DB, nil, h.TxID)
	u, err := helpers.GetUserContextV2(c)
	if err != nil {
		logger.Error.Printf("couldn't get user token: %v", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(1, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientWallet := wallet_proto.NewWalletServicesWalletClient(connAuth)

	token := c.Get("Authorization")[7:]
	ctx := grpcMetadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resWallet, err := clientWallet.GetWalletByUserId(ctx, &wallet_proto.RequestGetWalletByUserId{UserId: u.ID})
	if err != nil {
		logger.Error.Printf("error conectando con el servicio wallet de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio wallet de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resWallet == nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain")
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resWallet.Error {
		logger.Error.Printf(resWallet.Msg)
		res.Code, res.Type, res.Msg = int(resWallet.Code), int(resWallet.Type), resWallet.Msg
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resWallet.Data == nil {
		logger.Error.Printf("El usuario no tiene ninguna wallet asociada")
		res.Code, res.Type, res.Msg = 22, 1, "El usuario no tiene ninguna wallet asociada"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	resPenalty, err := srvBk.SrvPenalty.GetAllPenaltyParticipantsByWalletID(resWallet.Data.Id)
	if err != nil {
		logger.Error.Printf("error trayendo la cantidad de faltas, error: %v", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(22, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = Wallet{
		Id:             resWallet.Data.Id,
		Mnemonic:       resWallet.Data.Mnemonic,
		RsaPublic:      resWallet.Data.Public,
		IpDevice:       resWallet.Data.IpDevice,
		StatusId:       resWallet.Data.StatusId,
		IdentityNumber: resWallet.Data.IdentityNumber,
		Faults:         len(resPenalty),
		CreatedAt:      resWallet.Data.CreatedAt,
		UpdatedAt:      resWallet.Data.UpdatedAt,
	}
	res.Code, res.Type, res.Msg = msg.GetByCode(29, h.DB, h.TxID)
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

func (h *handlerUser) GetAccountByWalletID(c *fiber.Ctx) error {
	res := resAccount{Error: true}
	walletId := c.Params("wallet")
	e := env.NewConfiguration()

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientAccount := accounting_proto.NewAccountingServicesAccountingClient(connAuth)

	token := c.Get("Authorization")[7:]
	ctx := grpcMetadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resWsAccount, err := clientAccount.GetAccountingByWalletById(ctx, &accounting_proto.RequestGetAccountingByWalletId{Id: walletId})
	if err != nil {
		logger.Error.Printf("error conectando con el servicio account de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio account de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resWsAccount == nil {
		logger.Error.Printf("error conectando con el servicio account de blockchain")
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio account de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resWsAccount.Error {
		logger.Error.Printf(resWsAccount.Msg)
		res.Code, res.Type, res.Msg = int(resWsAccount.Code), int(resWsAccount.Type), resWsAccount.Msg
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = Accounting{
		Id:       resWsAccount.Data.Id,
		IdWallet: resWsAccount.Data.IdWallet,
		Amount:   resWsAccount.Data.Amount,
	}
	res.Code, res.Type, res.Msg = msg.GetByCode(29, h.DB, h.TxID)
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

func (h *handlerUser) GetFreezeMoney(c *fiber.Ctx) error {

	res := resFreezeMoney{Error: true}
	walletId := c.Params("wallet")

	e := env.NewConfiguration()
	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientWallet := wallet_proto.NewWalletServicesWalletClient(connAuth)

	token := c.Get("Authorization")[7:]
	ctx := grpcMetadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resFrozen, err := clientWallet.GetFrozenMoney(ctx, &wallet_proto.RqGetFrozenMoney{WalletId: walletId})
	if err != nil {
		logger.Error.Printf("error conectando con el servicio account de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio account de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resFrozen == nil {
		logger.Error.Printf("error conectando con el servicio account de blockchain")
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio account de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resFrozen.Error {
		logger.Error.Printf(resFrozen.Msg)
		res.Code, res.Type, res.Msg = int(resFrozen.Code), int(resFrozen.Type), resFrozen.Msg
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = resFrozen.Data
	res.Code, res.Type, res.Msg = msg.GetByCode(29, h.DB, h.TxID)
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

func (h *handlerUser) CreateUser(c *fiber.Ctx) error {
	res := responseAnny{Error: true}
	request := requestCreateUser{}
	e := env.NewConfiguration()
	err := c.BodyParser(&request)
	if err != nil {
		logger.Error.Printf("couldn't bind model request: %v", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(22, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if request.Password != request.ConfirmPassword {
		res.Code, res.Type, res.Msg = 22, 1, "La contraseña no conincide con la contra"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(22, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientUser := users_proto.NewAuthServicesUsersClient(connAuth)

	resCreateUser, err := clientUser.CreateUser(context.Background(), &users_proto.UserRequest{
		Id:              uuid.New().String(),
		Nickname:        request.Nickname,
		Email:           request.Email,
		Password:        request.Password,
		ConfirmPassword: request.ConfirmPassword,
		Name:            request.Name,
		Lastname:        request.Lastname,
		IdType:          int32(request.IdType),
		IdNumber:        request.IdNumber,
		Cellphone:       request.Cellphone,
		BirthDate:       request.BirthDate,
	})
	if err != nil {
		logger.Error.Printf("No se pudo crear el usuario, error: %s", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(22, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resCreateUser == nil {
		logger.Error.Printf("No se pudo crear el usuario")
		res.Code, res.Type, res.Msg = msg.GetByCode(22, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resCreateUser.Error {
		logger.Error.Printf(resCreateUser.Msg)
		res.Code, res.Type, res.Msg = msg.GetByCode(22, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = "Usuario creado correctamente, se envió un correo de confirmación a su correo electrónico"
	res.Code, res.Type, res.Msg = msg.GetByCode(29, h.DB, h.TxID)
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

// active user godoc
// @Summary BLion Api consenso
// @Description Active User
// @Accept  json
// @Produce  json
// @Success 200 {object} responseAnny
// @Success 202 {object} requestActivateUser
// @Router /api/v1/user/active [post]
// @Authorization Bearer token
func (h *handlerUser) activateUser(c *fiber.Ctx) error {
	e := env.NewConfiguration()
	res := responseAnny{Error: true}

	m := requestActivateUser{}
	err := c.BodyParser(&m)
	if err != nil {
		logger.Error.Printf("couldn't bind model create wallets: %v", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(1, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientUser := users_proto.NewAuthServicesUsersClient(connAuth)

	token := c.Get("Authorization")[7:]
	ctx := grpcMetadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resActivate, err := clientUser.ActivateUser(ctx, &users_proto.ActivateUserRequest{Code: m.Code})
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resActivate == nil {
		logger.Error.Printf("No se pudo activar la cuenta del usuario: %v", err)
		res.Code, res.Type, res.Msg = 22, 1, "No se pudo activar la cuenta del usuario"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resActivate.Error {
		logger.Error.Printf(resActivate.Msg)
		res.Code, res.Type, res.Msg = int(resActivate.Code), int(resActivate.Type), resActivate.Msg
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = "Cuenta activada y lista para ser usada"
	res.Code, res.Type, res.Msg = msg.GetByCode(29, h.DB, h.TxID)
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

func (h *handlerUser) activateWallet(c *fiber.Ctx) error {
	e := env.NewConfiguration()
	res := responseActivateWallet{Error: true}

	m := requestActivateWallet{}
	err := c.BodyParser(&m)
	if err != nil {
		logger.Error.Printf("couldn't bind model create wallets: %v", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(1, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientWallet := wallet_proto.NewWalletServicesWalletClient(connAuth)

	token := c.Get("Authorization")[7:]
	ctx := grpcMetadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resActivate, err := clientWallet.ActivateWallet(ctx, &wallet_proto.RequestActivateWallet{Id: m.Id, Mnemonic: m.Mnemonic})
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resActivate == nil {
		logger.Error.Printf("No se pudo activar la cuenta del usuario: %v", err)
		res.Code, res.Type, res.Msg = 22, 1, "No se pudo activar la cuenta del usuario"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resActivate.Error {
		logger.Error.Printf(resActivate.Msg)
		res.Code, res.Type, res.Msg = int(resActivate.Code), int(resActivate.Type), resActivate.Msg
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = &Key{
		Public:   resActivate.Data.Key.Public,
		Private:  resActivate.Data.Key.Private,
		Mnemonic: resActivate.Data.Mnemonic,
	}
	res.Code, res.Type, res.Msg = 29, 29, "Su wallet se activo correctamente"
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

func (h *handlerUser) createWallet(c *fiber.Ctx) error {
	e := env.NewConfiguration()
	res := responseCreateWallet{Error: true}

	usr, err := helpers.GetUserContextV2(c)
	if err != nil {
		logger.Error.Printf("No se pudo obtener el usuario de la sesión: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "No se pudo obtener el usuario de la sesión"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientWallet := wallet_proto.NewWalletServicesWalletClient(connAuth)

	token := c.Get("Authorization")[7:]
	ctx := grpcMetadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resCreateWt, err := clientWallet.CreateWallet(ctx, &wallet_proto.RequestCreateWallet{UserId: usr.ID})
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resCreateWt == nil {
		logger.Error.Printf("No se pudo crear la wallet: %v", err)
		res.Code, res.Type, res.Msg = 22, 1, "No se pudo crear la wallet"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resCreateWt.Error {
		logger.Error.Printf(resCreateWt.Msg)
		res.Code, res.Type, res.Msg = int(resCreateWt.Code), int(resCreateWt.Type), resCreateWt.Msg
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = requestActivateWallet{
		Id:       resCreateWt.Data.Id,
		Mnemonic: resCreateWt.Data.Mnemonic,
	}
	res.Code, res.Type, res.Msg = 29, 29, "Se ha creado la wallet, se ha enviado un correo electrónico a su correo para activar la cuenta"
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

func (h *handlerUser) RequestChangePwd(c *fiber.Ctx) error {
	res := responseAnny{Error: true}
	e := env.NewConfiguration()
	request := ReqChangePwd{}
	err := c.BodyParser(&request)
	if err != nil {
		logger.Error.Printf("couldn't bind model create wallets: %v", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(1, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientUser := users_proto.NewAuthServicesUsersClient(connAuth)

	resActivate, err := clientUser.RequestChangePassword(context.Background(), &users_proto.RqChangePwd{
		Email:    request.Email,
		Nickname: request.Nickname,
	})
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resActivate == nil {
		logger.Error.Printf("No se pudo solicitar el cambio de contraseña: %v", err)
		res.Code, res.Type, res.Msg = 22, 1, "No se pudo solicitar el cambio de contraseña"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resActivate.Error {
		logger.Error.Printf(resActivate.Msg)
		res.Code, res.Type, res.Msg = int(resActivate.Code), int(resActivate.Type), resActivate.Msg
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = "Se ha enviado un correo para que pueda restablecer su contraseña"
	res.Code, res.Type, res.Msg = msg.GetByCode(29, h.DB, h.TxID)
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}

func (h *handlerUser) ChangePassword(c *fiber.Ctx) error {
	res := responseAnny{Error: true}
	e := env.NewConfiguration()
	request := ChangePwd{}
	err := c.BodyParser(&request)
	if err != nil {
		logger.Error.Printf("couldn't bind model create wallets: %v", err)
		res.Code, res.Type, res.Msg = msg.GetByCode(1, h.DB, h.TxID)
		return c.Status(http.StatusAccepted).JSON(res)
	}

	connAuth, err := grpc.Dial(e.AuthService.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}
	defer connAuth.Close()

	clientUser := users_proto.NewAuthServicesUsersClient(connAuth)

	token := c.Get("Authorization")[7:]
	ctx := grpcMetadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resActivate, err := clientUser.ChangePassword(ctx, &users_proto.RequestChangePwd{
		OldPassword:     request.OldPassword,
		NewPassword:     request.NewPassword,
		ConfirmPassword: request.ConfirmPassword,
	})
	if err != nil {
		logger.Error.Printf("error conectando con el servicio auth de blockchain: %s", err)
		res.Code, res.Type, res.Msg = 22, 1, "error conectando con el servicio auth de blockchain"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resActivate == nil {
		logger.Error.Printf("No se pudo realizar el cambio de contraseña: %v", err)
		res.Code, res.Type, res.Msg = 22, 1, "No se pudo realizar el cambio de contraseña"
		return c.Status(http.StatusAccepted).JSON(res)
	}

	if resActivate.Error {
		logger.Error.Printf(resActivate.Msg)
		res.Code, res.Type, res.Msg = int(resActivate.Code), int(resActivate.Type), resActivate.Msg
		return c.Status(http.StatusAccepted).JSON(res)
	}

	res.Data = "Se ha cambiado correctamente la contraseña"
	res.Code, res.Type, res.Msg = msg.GetByCode(29, h.DB, h.TxID)
	res.Error = false
	return c.Status(http.StatusOK).JSON(res)
}
