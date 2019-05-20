package bankx

import (
	"github.com/coinexchain/dex/denoms"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/coinexchain/dex/x/authx"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case bank.MsgSend:
			return handleMsgSend(ctx, k, msg)
		default:
			errMsg := "Unrecognized bank Msg type: %s" + msg.Type()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}

}

func handleMsgSend(ctx sdk.Context, k Keeper, msg bank.MsgSend) sdk.Result {

	if !k.bk.GetSendEnabled(ctx) {
		return bank.ErrSendDisabled(k.bk.Codespace()).Result()
	}

	_, ok := k.axk.GetAccountX(ctx, msg.ToAddress)
	amt := msg.Amount

	var neg bool

	activatedFee := k.GetParam(ctx).ActivatedFee
	//toaccount doesn't exist yet
	if !ok {

		//check whether the first transfer contains at least activatedFee cet
		amt, neg = amt.SafeSub(denoms.NewCetCoins(activatedFee))

		if neg {
			return sdk.ErrInvalidCoins(amt.String()).Result()
		}

	}

	//handle coins transfer
	t, err := k.bk.SendCoins(ctx, msg.FromAddress, msg.ToAddress, amt)

	if err != nil {
		return err.Result()
	}

	// new accountx for toaddress if needed
	if !ok {
		newAccountX := authx.NewAccountXWithAddress(msg.ToAddress)
		newAccountX.Activated = true
		k.axk.SetAccountX(ctx, newAccountX)

		//collect account activation fees
		k.fck.AddCollectedFees(ctx, denoms.NewCetCoins(activatedFee))

		// sub the activatedfees from fromaddress
		fromAccount := k.ak.GetAccount(ctx, msg.FromAddress)
		oldCoins := fromAccount.GetCoins()
		newCoins, _ := oldCoins.SafeSub(denoms.NewCetCoins(activatedFee))
		//newCoins, _ := subActivatedFee(oldCoins, activatedFee)
		fromAccount.SetCoins(newCoins)
		k.ak.SetAccount(ctx, fromAccount)
	}

	return sdk.Result{
		Tags: t,
	}
}
