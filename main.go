package main

import (
	"os/exec"
	"strings"

	"syscall"

	"github.com/dontpanic92/wxGo/wx"
	"golang.org/x/sys/windows/registry"
)

type ControlDialog struct {
	wx.Dialog
}

type Account struct {
	label    string
	username string
	password string
}

type AccountListBox struct {
	wx.ListBox
	accountList []Account
}

var accListBox AccountListBox

func (aListBox *AccountListBox) UpdateList() {
	listBoxStrings := []string{}

	for i := 0; i < len(aListBox.accountList); i++ {
		listBoxStrings = append(listBoxStrings, aListBox.accountList[i].label+" ("+aListBox.accountList[i].username+")")
	}

	aListBox.Set(listBoxStrings)
}

func (aListBox *AccountListBox) AddAccount(acc Account) {
	aListBox.accountList = append(aListBox.accountList, acc)
	aListBox.UpdateList()
}

func (aListBox *AccountListBox) RemoveAccount(username string) {
	sliceRemove := func(slice []Account, s int) []Account { return append(slice[:s], slice[s+1:]...) }

	for i := 0; i < len(aListBox.accountList); i++ {
		if strings.ToLower(aListBox.accountList[i].username) == strings.ToLower(username) {
			aListBox.accountList = sliceRemove(aListBox.accountList, i)
			break
		}
	}

	aListBox.UpdateList()
}

func (aListBox *AccountListBox) RemoveAccountId(index int) {
	sliceRemove := func(slice []Account, s int) []Account { return append(slice[:s], slice[s+1:]...) }

	aListBox.accountList = sliceRemove(aListBox.accountList, index)
	aListBox.UpdateList()
}

func (aListBox *AccountListBox) UpdateAccount(index int, new Account) {
	aListBox.accountList[index] = new
	aListBox.UpdateList()
}

func ShowAccountDialog(editId int) {
	f := ControlDialog{}
	f.Dialog = wx.NewDialog(wx.NullWindow, wx.ID_ANY, "Add account...", wx.DefaultPosition, wx.NewSizeT(250, 180))

	sizerVert := wx.NewBoxSizer(wx.VERTICAL)

	sizerVert.AddSpacer(10)

	sizerHorzLabelInput := wx.NewBoxSizer(wx.HORIZONTAL)
	textLabel := wx.NewStaticText(f, wx.ID_ANY, "Label: ", wx.DefaultPosition, wx.DefaultSize, 0)
	textBoxLabel := wx.NewTextCtrl(f, wx.ID_ANY, "", wx.DefaultPosition, wx.DefaultSize, 0)

	sizerHorzLabelInput.Add(textLabel, 0, wx.ALL, 5)
	sizerHorzLabelInput.Add(textBoxLabel, 1, wx.ALL, 5)

	sizerVert.Add(sizerHorzLabelInput, 0, wx.LEFT|wx.RIGHT|wx.EXPAND, 20)

	sizerHorzUsernameInput := wx.NewBoxSizer(wx.HORIZONTAL)
	textUsername := wx.NewStaticText(f, wx.ID_ANY, "Username: ", wx.DefaultPosition, wx.DefaultSize, 0)
	textBoxUsername := wx.NewTextCtrl(f, wx.ID_ANY, "", wx.DefaultPosition, wx.DefaultSize, 0)

	sizerHorzUsernameInput.Add(textUsername, 0, wx.ALL, 5)
	sizerHorzUsernameInput.Add(textBoxUsername, 1, wx.ALL, 5)

	sizerVert.Add(sizerHorzUsernameInput, 0, wx.LEFT|wx.RIGHT|wx.EXPAND, 20)

	sizerHorzPasswordInput := wx.NewBoxSizer(wx.HORIZONTAL)
	textPassword := wx.NewStaticText(f, wx.ID_ANY, "Password: ", wx.DefaultPosition, wx.DefaultSize, 0)
	textBoxPassword := wx.NewTextCtrl(f, wx.ID_ANY, "", wx.DefaultPosition, wx.DefaultSize, 0)

	sizerHorzPasswordInput.Add(textPassword, 0, wx.ALL, 5)
	sizerHorzPasswordInput.Add(textBoxPassword, 1, wx.ALL, 5)

	sizerVert.Add(sizerHorzPasswordInput, 0, wx.LEFT|wx.RIGHT|wx.EXPAND, 20)

	buttonAdd := wx.NewButton(f, wx.ID_ANY, "Save", wx.DefaultPosition, wx.DefaultSize, 0)
	sizerVert.Add(buttonAdd, 0, wx.LEFT|wx.RIGHT|wx.EXPAND, 20)

	if editId != -1 {
		editAcc := accListBox.accountList[editId]
		textBoxLabel.SetValue(editAcc.label)
		textBoxUsername.SetValue(editAcc.username)
		textBoxPassword.SetValue(editAcc.password)
	}

	wx.Bind(f, wx.EVT_BUTTON, func(e2 wx.Event) {
		label := textBoxLabel.GetValue()
		username := strings.TrimSpace(textBoxUsername.GetValue())
		password := textBoxPassword.GetValue()

		if len(username) <= 0 {
			wx.MessageBox("Username cannot be empty", "Error", wx.ICON_ERROR)
			return
		} else if len(password) <= 0 {
			wx.MessageBox("Password cannot be empty", "Error", wx.ICON_ERROR)
			return
		}

		acc := Account{
			label:    label,
			username: username,
			password: password,
		}

		if editId != -1 {
			accListBox.UpdateAccount(editId, acc)
		} else {
			accListBox.AddAccount(acc)
		}
		f.Close(false)
	}, buttonAdd.GetId())

	f.SetSizer(sizerVert)
	f.Layout()
	f.Centre(wx.BOTH)
	f.ShowModal()
	f.Destroy()
}

func AddButtonHandler(e wx.Event) {
	ShowAccountDialog(-1)
}

func RemoveButtonHandler(e wx.Event) {
	if accListBox.GetSelection() == wx.NOT_FOUND {
		wx.MessageBox("Please select an account", "Error", wx.ICON_EXCLAMATION)
		return
	}

	accListBox.RemoveAccountId(accListBox.GetSelection())
}

func EditButtonHandler(e wx.Event) {
	if accListBox.GetSelection() == wx.NOT_FOUND {
		wx.MessageBox("Please select an account", "Error", wx.ICON_EXCLAMATION)
		return
	}

	ShowAccountDialog(accListBox.GetSelection())
}

//HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\Valve\Steam

func FindSteamInstallDir() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\WOW6432Node\\Valve\\Steam", registry.QUERY_VALUE)
	if err != nil {
		wx.MessageBox("Cannot find Steam! ("+err.Error()+")", "Error", wx.ICON_ERROR)
		return ""
	}

	val, _, err := k.GetStringValue("InstallPath")
	if err != nil {
		wx.MessageBox("Cannot find Steam! ("+err.Error()+")", "Error", wx.ICON_ERROR)
	}

	return val
}

func LoginButtonHandler(e wx.Event) {
	if accListBox.GetSelection() == wx.NOT_FOUND {
		wx.MessageBox("Please select an account", "Error", wx.ICON_EXCLAMATION)
		return
	}

	sel := accListBox.GetSelection()

	path := FindSteamInstallDir() + "\\steam.exe"
	cmd := exec.Command(path)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.CmdLine = `"` + path + `" -login ` + accListBox.accountList[sel].username + ` ` + accListBox.accountList[sel].password

	wx.MessageBox(cmd.String())
	err := cmd.Start()

	if err != nil {
		wx.MessageBox("Error while launching Steam!", "Error", wx.ICON_ERROR)
		return
	}
}

func OptionsButtonHandler(e wx.Event) {

}

func NewControlDialog() ControlDialog {
	f := ControlDialog{}
	f.Dialog = wx.NewDialog(wx.NullWindow, -1, "go-steamaccmgr", wx.DefaultPosition, wx.NewSizeT(300, 540))

	sizerVert := wx.NewBoxSizer(wx.VERTICAL)
	sizerVert.AddSpacer(15)
	sizerVert.Add(wx.NewStaticText(f, wx.ID_ANY, "Steam Account Manager v0.1", wx.DefaultPosition, wx.DefaultSize, 0),
		0, wx.ALL|wx.ALIGN_CENTRE_HORIZONTAL, 0)
	sizerVert.Add(wx.NewStaticText(f, wx.ID_ANY, "github.com/tax1driver", wx.DefaultPosition, wx.DefaultSize, 0),
		0, wx.ALL|wx.ALIGN_CENTRE_HORIZONTAL, 0)
	sizerVert.AddSpacer(15)

	accListBox = AccountListBox{}
	accListBox.ListBox = wx.NewListBox(f, wx.ID_ANY, wx.DefaultPosition, wx.NewSizeT(225, 300), []string{}, 0)
	sizerVert.Add(accListBox, 0, wx.LEFT|wx.RIGHT, 40)

	sizerVert.AddSpacer(5)

	sizerHorz := wx.NewBoxSizer(wx.HORIZONTAL)
	btnAddAccount := wx.NewButton(f, wx.ID_ANY, "Add", wx.DefaultPosition, wx.NewSizeT(75, -1), 0)
	btnRemoveAccount := wx.NewButton(f, wx.ID_ANY, "Remove", wx.DefaultPosition, wx.NewSizeT(75, -1), 0)
	btnEditAccount := wx.NewButton(f, wx.ID_ANY, "Edit", wx.DefaultPosition, wx.NewSizeT(75, -1), 0)

	sizerHorz.Add(btnAddAccount, 0, 0, 0)
	sizerHorz.Add(btnRemoveAccount, 0, 0, 0)
	sizerHorz.Add(btnEditAccount, 0, 0, 0)

	sizerVert.Add(sizerHorz, 0, wx.LEFT|wx.RIGHT, 39)

	btnLogin := wx.NewButton(f, wx.ID_ANY, "Login", wx.DefaultPosition, wx.DefaultSize, 0)
	sizerVert.Add(btnLogin, 0, wx.EXPAND|wx.LEFT|wx.RIGHT, 39)
	sizerVert.AddSpacer(15)

	btnOptions := wx.NewButton(f, wx.ID_ANY, "Options...", wx.DefaultPosition, wx.DefaultSize, 0)
	sizerVert.Add(btnOptions, 0, wx.EXPAND|wx.LEFT|wx.RIGHT, 39)

	// Bind handlers.
	wx.Bind(f, wx.EVT_BUTTON, AddButtonHandler, btnAddAccount.GetId())
	wx.Bind(f, wx.EVT_BUTTON, RemoveButtonHandler, btnRemoveAccount.GetId())
	wx.Bind(f, wx.EVT_BUTTON, EditButtonHandler, btnEditAccount.GetId())
	wx.Bind(f, wx.EVT_BUTTON, LoginButtonHandler, btnLogin.GetId())
	wx.Bind(f, wx.EVT_BUTTON, OptionsButtonHandler, btnOptions.GetId())

	f.SetSizer(sizerVert)
	f.Layout()
	f.Centre(wx.BOTH)

	return f
}

func main() {
	wx.NewApp()
	f := NewControlDialog()
	f.ShowModal()
	f.Destroy()
	return
}
