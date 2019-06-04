package asset

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestTokenKeeper_IssueToken(t *testing.T) {
	input := setupTestInput()

	type args struct {
		ctx sdk.Context
		msg MsgIssueToken
	}
	tests := []struct {
		name string
		args args
		want sdk.Error
	}{
		{
			"base-case",
			args{
				input.ctx,
				NewMsgIssueToken("ABC Token", "abc", 2100, tAccAddr,
					false, false, false, false),
			},
			nil,
		},
		{
			"case-duplicate",
			args{
				input.ctx,
				NewMsgIssueToken("ABC Token", "abc", 2100, tAccAddr,
					false, false, false, false),
			},
			ErrorDuplicateTokenSymbol("token symbol already exists in store"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := input.tk.IssueToken(tt.args.ctx, tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TokenKeeper.IssueToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenKeeper_TokenStore(t *testing.T) {
	input := setupTestInput()

	// set token
	token1, err := NewToken("ABC token", "abc", 2100, tAccAddr,
		false, false, false, false)
	require.NoError(t, err)
	err = input.tk.setToken(input.ctx, token1)
	require.NoError(t, err)

	token2, err := NewToken("XYZ token", "xyz", 2100, tAccAddr,
		false, false, false, false)
	require.NoError(t, err)
	err = input.tk.setToken(input.ctx, token2)
	require.NoError(t, err)

	// get all tokens
	tokens := input.tk.GetAllTokens(input.ctx)
	require.Equal(t, 2, len(tokens))
	require.Contains(t, []string{"abc", "xyz"}, tokens[0].GetSymbol())
	require.Contains(t, []string{"abc", "xyz"}, tokens[1].GetSymbol())

	// remove token
	input.tk.removeToken(input.ctx, token1)

	// get token
	res := input.tk.GetToken(input.ctx, token1.GetSymbol())
	require.Nil(t, res)

}

func TestTokenKeeper_TransferOwnership(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"
	var addr1, _ = sdk.AccAddressFromBech32("cosmos1n9e8krs6dengw6k8ts0xpntyzd27rhj48ve5gd")

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		false, false, false, false)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	msg := NewMsgTransferOwnership(symbol, tAccAddr, addr1)
	err = input.tk.TransferOwnership(input.ctx, msg)
	require.NoError(t, err)

	// get token
	token := input.tk.GetToken(input.ctx, symbol)
	require.NotNil(t, token)
	require.Equal(t, addr1.String(), token.GetOwner().String())

	//case2: invalid token
	msg = NewMsgTransferOwnership("xyz", tAccAddr, addr1)
	err = input.tk.TransferOwnership(input.ctx, msg)
	require.Error(t, err)

	//case3: invalid original owner
	msg = NewMsgTransferOwnership(symbol, tAccAddr, addr1)
	err = input.tk.TransferOwnership(input.ctx, msg)
	require.Error(t, err)

	//case4: invalid new owner
	msg = NewMsgTransferOwnership(symbol, addr1, sdk.AccAddress{})
	err = input.tk.TransferOwnership(input.ctx, msg)
	require.Error(t, err)
}

func TestTokenKeeper_MintToken(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"
	var addr, _ = sdk.AccAddressFromBech32("cosmos1n9e8krs6dengw6k8ts0xpntyzd27rhj48ve5gd")

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, false, false, false)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	msg := NewMsgMintToken(symbol, 1000, tAccAddr)
	err = input.tk.MintToken(input.ctx, msg)
	require.NoError(t, err)

	token := input.tk.GetToken(input.ctx, symbol)
	require.Equal(t, int64(3100), token.GetTotalSupply())
	require.Equal(t, int64(1000), token.GetTotalMint())

	err = input.tk.MintToken(input.ctx, msg)
	require.NoError(t, err)
	token = input.tk.GetToken(input.ctx, symbol)
	require.Equal(t, int64(4100), token.GetTotalSupply())
	require.Equal(t, int64(2000), token.GetTotalMint())

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 2: un mintable token
	// set token mintable: false
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		false, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	msg = NewMsgMintToken(symbol, 1000, tAccAddr)
	err = input.tk.MintToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 3: mint invalid token
	issueMsg = NewMsgIssueToken("ABC token", "xyz", 2100, tAccAddr,
		true, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgMintToken(symbol, 1000, tAccAddr)
	err = input.tk.MintToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 4: only token owner can mint token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, addr,
		true, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgMintToken(symbol, 1000, tAccAddr)
	err = input.tk.MintToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 5: token total mint amt is invalid
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgMintToken(symbol, 9E18+1, tAccAddr)
	err = input.tk.MintToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 6: token total supply before 1e8 boosting should be less than 90 billion
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgMintToken(symbol, 9E18, tAccAddr)
	err = input.tk.MintToken(input.ctx, msg)
	require.Error(t, err)
}

func TestTokenKeeper_BurnToken(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"
	var addr, _ = sdk.AccAddressFromBech32("cosmos1n9e8krs6dengw6k8ts0xpntyzd27rhj48ve5gd")

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, false)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	msg := NewMsgBurnToken(symbol, 1000, tAccAddr)
	err = input.tk.BurnToken(input.ctx, msg)
	require.NoError(t, err)

	token := input.tk.GetToken(input.ctx, symbol)
	require.Equal(t, int64(1100), token.GetTotalSupply())
	require.Equal(t, int64(1000), token.GetTotalBurn())

	err = input.tk.BurnToken(input.ctx, msg)
	require.NoError(t, err)
	token = input.tk.GetToken(input.ctx, symbol)
	require.Equal(t, int64(100), token.GetTotalSupply())
	require.Equal(t, int64(2000), token.GetTotalBurn())

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 2: un burnable token
	// set token burnable: false
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		false, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	msg = NewMsgBurnToken(symbol, 1000, tAccAddr)
	err = input.tk.BurnToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 3: burn invalid token
	issueMsg = NewMsgIssueToken("ABC token", "xyz", 2100, tAccAddr,
		true, true, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgBurnToken(symbol, 1000, tAccAddr)
	err = input.tk.BurnToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 4: only token owner can burn token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, addr,
		true, true, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgBurnToken(symbol, 1000, tAccAddr)
	err = input.tk.BurnToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 5: token total burn amt is invalid
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgBurnToken(symbol, 9E18+1, tAccAddr)
	err = input.tk.BurnToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 6: token total supply limited to > 0
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgBurnToken(symbol, 2100, tAccAddr)
	err = input.tk.BurnToken(input.ctx, msg)
	require.Error(t, err)
}

func TestTokenKeeper_ForbidToken(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"
	var addr, _ = sdk.AccAddressFromBech32("cosmos1n9e8krs6dengw6k8ts0xpntyzd27rhj48ve5gd")

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, true)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	msg := NewMsgForbidToken(symbol, tAccAddr)
	err = input.tk.ForbidToken(input.ctx, msg)
	require.NoError(t, err)

	token := input.tk.GetToken(input.ctx, symbol)
	require.Equal(t, true, token.GetIsForbidden())

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 2: un forbiddable token
	// set token forbiddable: false
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		false, false, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	msg = NewMsgForbidToken(symbol, tAccAddr)
	err = input.tk.ForbidToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 3: duplicate forbid token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, true)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgForbidToken(symbol, tAccAddr)
	err = input.tk.ForbidToken(input.ctx, msg)
	require.NoError(t, err)

	err = input.tk.ForbidToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 4: only token owner can forbid token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, addr,
		true, true, false, true)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	msg = NewMsgForbidToken(symbol, tAccAddr)
	err = input.tk.ForbidToken(input.ctx, msg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

}

func TestTokenKeeper_UnForbidToken(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, true)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	forbidMsg := NewMsgForbidToken(symbol, tAccAddr)
	err = input.tk.ForbidToken(input.ctx, forbidMsg)
	require.NoError(t, err)

	token := input.tk.GetToken(input.ctx, symbol)
	require.Equal(t, true, token.GetIsForbidden())

	unforbidMsg := NewMsgUnForbidToken(symbol, tAccAddr)
	err = input.tk.UnForbidToken(input.ctx, unforbidMsg)
	require.NoError(t, err)

	token = input.tk.GetToken(input.ctx, symbol)
	require.Equal(t, false, token.GetIsForbidden())

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 2: unforbid token before forbid token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, true)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	unforbidMsg = NewMsgUnForbidToken(symbol, tAccAddr)
	err = input.tk.UnForbidToken(input.ctx, unforbidMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)
}

func TestTokenKeeper_AddTokenWhitelist(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"
	whitelist := mockWhitelist()

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, true)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	token := input.tk.GetToken(input.ctx, symbol)

	addMsg := NewMsgAddTokenWhitelist(symbol, tAccAddr, whitelist)
	err = input.tk.AddTokenWhitelist(input.ctx, addMsg)
	require.NoError(t, err)
	addresses := input.tk.GetWhitelist(input.ctx, symbol)
	for _, addr := range addresses {
		require.Contains(t, whitelist, addr)
	}
	require.Equal(t, len(whitelist), len(addresses))

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 2: un forbiddable token
	// set token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	addMsg = NewMsgAddTokenWhitelist(symbol, tAccAddr, whitelist)
	err = input.tk.AddTokenWhitelist(input.ctx, addMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 3: nil whitelist
	// set token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, true)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	addMsg = NewMsgAddTokenWhitelist(symbol, tAccAddr, []sdk.AccAddress{})
	err = input.tk.AddTokenWhitelist(input.ctx, addMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)
}

func TestTokenKeeper_RemoveTokenWhitelist(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"
	whitelist := mockWhitelist()

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, true)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	token := input.tk.GetToken(input.ctx, symbol)

	addMsg := NewMsgAddTokenWhitelist(symbol, tAccAddr, whitelist)
	err = input.tk.AddTokenWhitelist(input.ctx, addMsg)
	require.NoError(t, err)
	addresses := input.tk.GetWhitelist(input.ctx, symbol)
	for _, addr := range addresses {
		require.Contains(t, whitelist, addr)
	}
	require.Equal(t, len(whitelist), len(addresses))

	removeMsg := NewMsgRemoveTokenWhitelist(symbol, tAccAddr, []sdk.AccAddress{whitelist[0]})
	err = input.tk.RemoveTokenWhitelist(input.ctx, removeMsg)
	require.NoError(t, err)
	addresses = input.tk.GetWhitelist(input.ctx, symbol)
	require.Equal(t, len(whitelist)-1, len(addresses))
	require.NotContains(t, addresses, whitelist[0])

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 2: un-forbiddable token
	// set token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	removeMsg = NewMsgRemoveTokenWhitelist(symbol, tAccAddr, whitelist)
	err = input.tk.RemoveTokenWhitelist(input.ctx, removeMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 3: nil whitelist
	// set token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, true)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	removeMsg = NewMsgRemoveTokenWhitelist(symbol, tAccAddr, []sdk.AccAddress{})
	err = input.tk.RemoveTokenWhitelist(input.ctx, removeMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)
}

func TestTokenKeeper_ForbidAddress(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"
	mock := mockAddresses()

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, true, true)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	token := input.tk.GetToken(input.ctx, symbol)

	forbidMsg := NewMsgForbidAddr(symbol, tAccAddr, mock)
	err = input.tk.ForbidAddress(input.ctx, forbidMsg)
	require.NoError(t, err)
	forbidden := input.tk.GetForbiddenAddr(input.ctx, symbol)
	for _, addr := range forbidden {
		require.Contains(t, mock, addr)
	}
	require.Equal(t, len(mock), len(forbidden))

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 2: addr un-forbiddable token
	// set token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	forbidMsg = NewMsgForbidAddr(symbol, tAccAddr, mock)
	err = input.tk.ForbidAddress(input.ctx, forbidMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 3: nil forbid address
	// set token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, true, true)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	forbidMsg = NewMsgForbidAddr(symbol, tAccAddr, []sdk.AccAddress{})
	err = input.tk.ForbidAddress(input.ctx, forbidMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)
}

func TestTokenKeeper_UnForbidAddress(t *testing.T) {
	input := setupTestInput()
	symbol := "abc"
	mock := mockAddresses()

	//case 1: base-case ok
	// set token
	issueMsg := NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, true, true)
	err := input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)
	token := input.tk.GetToken(input.ctx, symbol)

	forbidMsg := NewMsgForbidAddr(symbol, tAccAddr, mock)
	err = input.tk.ForbidAddress(input.ctx, forbidMsg)
	require.NoError(t, err)
	forbidden := input.tk.GetForbiddenAddr(input.ctx, symbol)
	for _, addr := range forbidden {
		require.Contains(t, mock, addr)
	}
	require.Equal(t, len(mock), len(forbidden))

	unForbidMsg := NewMsgUnForbidAddr(symbol, tAccAddr, []sdk.AccAddress{mock[0]})
	err = input.tk.UnForbidAddress(input.ctx, unForbidMsg)
	require.NoError(t, err)
	forbidden = input.tk.GetForbiddenAddr(input.ctx, symbol)
	require.Equal(t, len(mock)-1, len(forbidden))
	require.NotContains(t, forbidden, mock[0])

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 2: addr un-forbiddable token
	// set token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, false, false)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	unForbidMsg = NewMsgUnForbidAddr(symbol, tAccAddr, mock)
	err = input.tk.UnForbidAddress(input.ctx, unForbidMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)

	//case 3: nil un-forbid address
	// set token
	issueMsg = NewMsgIssueToken("ABC token", symbol, 2100, tAccAddr,
		true, true, true, true)
	err = input.tk.IssueToken(input.ctx, issueMsg)
	require.NoError(t, err)

	unForbidMsg = NewMsgUnForbidAddr(symbol, tAccAddr, []sdk.AccAddress{})
	err = input.tk.UnForbidAddress(input.ctx, unForbidMsg)
	require.Error(t, err)

	// remove token
	input.tk.removeToken(input.ctx, token)
}
