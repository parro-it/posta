package imap

import (
	"testing"
	"time"

	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var a = Account{
	Cfg: config.Account{
		Addr:     "test.mailu.io:143",
		User:     "admin@test.mailu.io",
		Pass:     "letmein",
		Name:     "test mailbox",
		StartTLS: true,
	},
	client: nil,
}

func TestLogin(t *testing.T) {

	res := a.Login()
	<-res.Res
	assert.NoError(t, res.Err)
}

func TestListFolders(t *testing.T) {
	res := a.Login()
	<-res.Res
	require.NoError(t, res.Err)
	fold := a.ListFolders()
	arr := chans.Collect(fold.Res)
	require.NoError(t, fold.Err)
	for i := range arr {
		require.NotNil(t, arr[i].mbInfo)
		arr[i].mbInfo = nil
	}
	assert.Equal(t, []Folder{
		{Size: 0x0, Name: "Sent", Account: "test mailbox", Path: "Sent"},
		{Size: 0x0, Name: "Trash", Account: "test mailbox", Path: "Trash"},
		{Size: 0x0, Name: "Junk", Account: "test mailbox", Path: "Junk"},
		{Size: 0x0, Name: "Drafts", Account: "test mailbox", Path: "Drafts"},
		{Size: 0x2, Name: "INBOX", Account: "test mailbox", Path: "INBOX"},
	}, arr)

}
func TestListMessages(t *testing.T) {
	res := a.Login()
	<-res.Res
	require.NoError(t, res.Err)
	fold := a.ListFolders()
	var inbox Folder
	for fo := range fold.Res {
		if fo.Name == "INBOX" {
			inbox = fo
			break
		}
	}
	require.NoError(t, fold.Err)

	mes := a.ListMessages(inbox)
	msgs := chans.Collect(mes.Res)
	require.NoError(t, mes.Err)
	/*for i := range arr {
		require.NotNil(t, arr[i].mbInfo)
		arr[i].mbInfo = nil
	}*/

	assert.Equal(t, []Msg{
		{Date: time.Date(2022, time.January, 18, 11, 21, 54, 0, msgs[0].Date.Location()), From: "Mail Delivery System", To: []string{""}, Subject: "Delayed Mail (still being retried)"},
		{Date: time.Date(2022, time.January, 18, 11, 41, 19, 0, msgs[1].Date.Location()), From: "Mail Delivery System", To: []string{""}, Subject: "Delayed Mail (still being retried)"}}, msgs)

}
