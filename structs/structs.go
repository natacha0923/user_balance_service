package structs

type UserBalance struct {
	UserId  uint64
	Balance int64
}

func (balance *UserBalance) IsEmpty() bool {
	return balance.UserId == 0
}

type ChangeBalanceRequest struct {
	Token string                   `json:"token"`
	Body  ChangeBalanceRequestBody `json:"body"`
}

type ChangeBalanceRequestBody struct {
	UserId uint64 `json:"userId" validate:"min=1"`
	Amount int64  `json:"amount" validate:"nonzero"`
}

type TransferRequest struct {
	Token string              `json:"token"`
	Body  TransferRequestBody `json:"body"`
}

type TransferRequestBody struct {
	From   uint64 `json:"from" validate:"min=1"`
	To     uint64 `json:"to" validate:"min=1"`
	Amount int64  `json:"amount" validate:"min=1"`
}
