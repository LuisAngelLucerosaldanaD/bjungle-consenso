package user

type rqLogin struct {
	NickName string `json:"nick_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type responseLogin struct {
	Error bool   `json:"error"`
	Data  Token  `json:"data"`
	Code  int    `json:"code"`
	Type  int    `json:"type"`
	Msg   string `json:"msg"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type resGetWallets struct {
	Error bool   `json:"error"`
	Data  Wallet `json:"data"`
	Code  int    `json:"code"`
	Type  int    `json:"type"`
	Msg   string `json:"msg"`
}

type Wallet struct {
	Id             string `json:"id,omitempty"`
	Mnemonic       string `json:"mnemonic,omitempty"`
	RsaPublic      string `json:"rsa_public,omitempty"`
	IpDevice       string `json:"ip_device,omitempty"`
	StatusId       int32  `json:"status_id,omitempty"`
	IdentityNumber string `json:"identity_number,omitempty"`
	Faults         int    `json:"faults"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

type resAccount struct {
	Error bool       `json:"error"`
	Data  Accounting `json:"data"`
	Code  int        `json:"code"`
	Type  int        `json:"type"`
	Msg   string     `json:"msg"`
}

type Accounting struct {
	Id        string  `json:"id,omitempty"`
	IdWallet  string  `json:"id_wallet,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	IdUser    string  `json:"id_user,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

type resFreezeMoney struct {
	Error bool    `json:"error"`
	Data  float64 `json:"data"`
	Code  int     `json:"code"`
	Type  int     `json:"type"`
	Msg   string  `json:"msg"`
}

type responseAnny struct {
	Error bool   `json:"error"`
	Data  string `json:"data"`
	Code  int    `json:"code"`
	Type  int    `json:"type"`
	Msg   string `json:"msg"`
}

type responseActivateWallet struct {
	Error bool   `json:"error"`
	Data  *Key   `json:"data"`
	Code  int    `json:"code"`
	Type  int    `json:"type"`
	Msg   string `json:"msg"`
}

type Key struct {
	Public   string `json:"public"`
	Private  string `json:"private"`
	Mnemonic string `json:"mnemonic"`
}

type responseCreateWallet struct {
	Error bool                  `json:"error"`
	Data  requestActivateWallet `json:"data"`
	Code  int                   `json:"code"`
	Type  int                   `json:"type"`
	Msg   string                `json:"msg"`
}

type requestActivateUser struct {
	Code string `json:"code"`
}

type requestActivateWallet struct {
	Id       string `json:"id"`
	Mnemonic string `json:"mnemonic"`
}

type ReqChangePwd struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type ChangePwd struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type requestCreateUser struct {
	Nickname        string `json:"nickname"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Name            string `json:"name"`
	Lastname        string `json:"lastname"`
	IdType          int    `json:"id_type"`
	IdNumber        string `json:"id_number"`
	Cellphone       string `json:"cellphone"`
	BirthDate       string `json:"birth_date"`
}
