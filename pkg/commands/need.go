package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
)

type needItemHandler struct {
	user     storage.User
	charName string
}

func (h needItemHandler) HandleLine(user string, args []rune) (string, error) {
	args, ctRunes := parser.MaybeCount(args)

	itemName := strings.TrimSpace(string(args))

	if len(itemName) == 0 {
		return "", ErrItemNameRequired
	}

	ctStr := strings.TrimSpace(string(ctRunes))
	if ctStr == "" {
		ctStr = "1"
	}

	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return "", errors.Wrap(err, "could not interpret count to adjust item needs")
	}

	if ct < 0 {
		return "", ErrPositiveValueRequired
	}

	char, err := h.user.GetCharacter(h.charName)
	if err != nil {
		return "", errors.Wrap(err, "could not find character to adjust item needs")
	}

	char.IncrNeededItem(itemName, uint64(ct))

	return fmt.Sprintf("marked %s as needing +%d of %s", h.charName, ct, itemName), nil
}

type needPointsHandler struct {
	user     storage.User
	charName string
}

func (h needPointsHandler) HandleLine(user string, args []rune) (string, error) {
	args, ctRunes := parser.MaybeCount(args)

	skillName := strings.TrimSpace(string(args))

	if len(skillName) == 0 {
		return "", ErrSkillNameRequired
	}

	ctStr := strings.TrimSpace(string(ctRunes))
	if ctStr == "" {
		ctStr = "1"
	}

	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return "", errors.Wrap(err, "could not interpret count to adjust skill needs")
	}

	if ct < 0 {
		return "", ErrPositiveValueRequired
	}

	char, err := h.user.GetCharacter(h.charName)
	if err != nil {
		return "", errors.Wrap(err, "could not find character to adjust skill needs")
	}

	char.IncrNeededSkill(skillName, uint64(ct))

	return fmt.Sprintf("marked %s as needing +%d points in %s", h.charName, ct, skillName), nil
}

type needCommands struct {
	preCommand string
	deps       dependencies
}

func (c needCommands) helpCharsPoints(user string, args []rune) (string, error) {
	helpStr := fmt.Sprintf("Usage: %s [%s] [skill name] [count?]\n\n", c.preCommand+" pts", "charname")
	helpStr += fmt.Sprintf("Available %ss:\n", "charnames")

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return helpStr, nil
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		return helpStr, nil
	}

	for _, char := range bUser.GetCharacters() {
		helpStr += fmt.Sprintf("  %s\n", char.GetName())
	}

	return helpStr, nil
}

func (c needCommands) helpCharsItems(user string, args []rune) (string, error) {
	helpStr := fmt.Sprintf("Usage: %s [%s] [item name] [count?]\n\n", c.preCommand+" pts", "charname")
	helpStr += fmt.Sprintf("Available %ss:\n", "charnames")

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return helpStr, nil
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		return helpStr, nil
	}

	for _, char := range bUser.GetCharacters() {
		helpStr += fmt.Sprintf("  %s\n", char.GetName())
	}

	return helpStr, nil
}

func (c needCommands) points(user string, args []rune) (string, error) {
	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
		if err != nil {
			return "", errors.Wrap(err, "could not create user")
		}
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, len(characters))
	for i, char := range characters {
		charNames[i] = char.GetName()
	}

	p := parser.NewParser(parser.Options{
		CmdIndicator:  ' ',
		KnownCommands: append(charNames, ""),
	})
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  c.preCommand + " pts",
		Placeholder: "charname",
	})
	ch.SetHandler("", cmdhandler.NewLineHandler(c.helpCharsPoints))
	for _, char := range characters {
		ch.SetHandler(char.GetName(), needPointsHandler{charName: char.GetName(), user: bUser})
	}
	retStr, err := ch.HandleLine(user, args)

	if err != nil {
		return retStr, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return "", errors.Wrap(err, "could not save points need")
	}

	err = t.Commit()
	if err != nil {
		return "", errors.Wrap(err, "could not save points need")
	}

	return retStr, nil
}

func (c needCommands) item(user string, args []rune) (string, error) {
	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
		if err != nil {
			return "", errors.Wrap(err, "could not create user")
		}
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, len(characters))
	for i, char := range characters {
		charNames[i] = char.GetName()
	}

	p := parser.NewParser(parser.Options{
		CmdIndicator:  ' ',
		KnownCommands: append(charNames, ""),
	})
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  c.preCommand + " item",
		Placeholder: "charname",
	})
	ch.SetHandler("", cmdhandler.NewLineHandler(c.helpCharsItems))
	for _, char := range characters {
		ch.SetHandler(char.GetName(), needItemHandler{charName: char.GetName(), user: bUser})
	}
	retStr, err := ch.HandleLine(user, args)

	if err != nil {
		return retStr, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return "", errors.Wrap(err, "could not save item need")
	}

	err = t.Commit()
	if err != nil {
		return "", errors.Wrap(err, "could not save item need")
	}

	return retStr, nil
}

// NeedCommandHandler TODOC
func NeedCommandHandler(deps dependencies, preCommand string) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: ' ',
		KnownCommands: []string{
			"help",
			"pts",
			"item",
		},
	})
	nc := needCommands{
		preCommand: preCommand,
		deps:       deps,
	}
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  preCommand,
		Placeholder: "type",
	})
	ch.SetHandler("pts", cmdhandler.NewLineHandler(nc.points))
	ch.SetHandler("item", cmdhandler.NewLineHandler(nc.item))

	return ch
}