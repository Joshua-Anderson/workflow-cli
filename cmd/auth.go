package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/controller-sdk-go/auth"
	"github.com/deis/workflow-cli/settings"
	"golang.org/x/crypto/ssh/terminal"
)

// Register creates a account on a Deis controller.
func Register(cf string, controller string, username string, password string, email string,
	sslVerify bool, wOut io.Writer) error {

	c, err := deis.New(sslVerify, controller, "")

	if err != nil {
		return err
	}

	tempSettings, err := settings.Load(cf)

	if err == nil && tempSettings.Client.ControllerURL.Host == c.ControllerURL.Host {
		c.Token = tempSettings.Client.Token
	}

	// Set user agent for temporary client.
	c.UserAgent = settings.UserAgent

	if err = c.CheckConnection(); checkAPICompatibility(c, err, wOut) != nil {
		return err
	}

	if username == "" {
		fmt.Fprint(wOut, "username: ")
		fmt.Scanln(&username)
	}

	if password == "" {
		fmt.Fprint(wOut, "password: ")
		password, err = readPassword()
		fmt.Fprintf(wOut, "\npassword (confirm): ")
		passwordConfirm, err := readPassword()
		fmt.Fprintln(wOut)

		if err != nil {
			return err
		}

		if password != passwordConfirm {
			return errors.New("Password mismatch, aborting registration.")
		}
	}

	if email == "" {
		fmt.Fprint(wOut, "email: ")
		fmt.Scanln(&email)
	}

	err = auth.Register(c, username, password, email)

	c.Token = ""

	if checkAPICompatibility(c, err, wOut) != nil {
		fmt.Fprint(os.Stderr, "Registration failed: ")
		return err
	}

	fmt.Fprintf(wOut, "Registered %s\n", username)

	s := settings.Settings{Client: c}
	return doLogin(cf, s, username, password, wOut)
}

func doLogin(cf string, s settings.Settings, username, password string, wOut io.Writer) error {
	token, err := auth.Login(s.Client, username, password)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	s.Client.Token = token
	s.Username = username

	filename, err := s.Save(cf)

	if err != nil {
		return nil
	}

	fmt.Fprintf(wOut, "Logged in as %s\n", username)
	fmt.Fprintf(wOut, "Configuration file written to %s\n", filename)
	return nil
}

// Login to a Deis controller.
func Login(cf string, controller string, username string, password string, sslVerify bool,
	wOut io.Writer) error {
	c, err := deis.New(sslVerify, controller, "")

	if err != nil {
		return err
	}

	// Set user agent for temporary client.
	c.UserAgent = settings.UserAgent

	if err = c.CheckConnection(); checkAPICompatibility(c, err, wOut) != nil {
		return err
	}

	if username == "" {
		fmt.Fprint(wOut, "username: ")
		fmt.Scanln(&username)
	}

	if password == "" {
		fmt.Fprint(wOut, "password: ")
		password, err = readPassword()
		fmt.Fprintln(wOut)

		if err != nil {
			return err
		}
	}

	s := settings.Settings{Client: c}
	return doLogin(cf, s, username, password, wOut)
}

// Logout from a Deis controller.
func Logout(cf string, wOut io.Writer) error {
	if err := settings.Delete(cf); err != nil {
		return err
	}

	fmt.Fprintln(wOut, "Logged out")
	return nil
}

// Passwd changes a user's password.
func Passwd(cf, username, password, newPassword string, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	if password == "" && username == "" {
		fmt.Fprint(wOut, "current password: ")
		password, err = readPassword()
		fmt.Fprintln(wOut)

		if err != nil {
			return err
		}
	}

	if newPassword == "" {
		fmt.Fprint(wOut, "new password: ")
		newPassword, err = readPassword()
		fmt.Fprintf(wOut, "\nnew password (confirm): ")
		passwordConfirm, err := readPassword()

		fmt.Fprintln(wOut)

		if err != nil {
			return err
		}

		if newPassword != passwordConfirm {
			return errors.New("Password mismatch, not changing.")
		}
	}

	err = auth.Passwd(s.Client, username, password, newPassword)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		fmt.Fprint(os.Stderr, "Password change failed: ")
		return err
	}

	fmt.Fprintln(wOut, "Password change succeeded.")
	return nil
}

// Cancel deletes a user's account.
func Cancel(cf, username, password string, yes bool, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	if username == "" || password != "" {
		fmt.Fprintln(wOut, "Please log in again in order to cancel this account")

		if err = Login(cf, s.Client.ControllerURL.String(), username, password, s.Client.VerifySSL, wOut); err != nil {
			return err
		}
	}

	if !yes {
		confirm := ""

		s, err = settings.Load(cf)

		if err != nil {
			return err
		}

		deletedUser := username

		if deletedUser == "" {
			deletedUser = s.Username
		}

		fmt.Fprintf(wOut, "cancel account %s at %s? (y/N): ", deletedUser, s.Client.ControllerURL.String())
		fmt.Scanln(&confirm)

		if strings.ToLower(confirm) == "y" {
			yes = true
		}
	}

	if !yes {
		fmt.Fprintln(os.Stderr, "Account not changed")
		return nil
	}

	err = auth.Delete(s.Client, username)
	if err == deis.ErrConflict {
		return fmt.Errorf("%s still has applications associated with it. Transfer ownership or delete them first", username)
	} else if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	// If user targets themselves, logout.
	if username == "" || s.Username == username {
		if err := settings.Delete(cf); err != nil {
			return err
		}
	}

	fmt.Fprintln(wOut, "Account cancelled")
	return nil
}

// Whoami prints the logged in user. If all is true, it fetches info from the controller to know
// more about the user.
func Whoami(cf string, all bool, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	if all {
		user, err := auth.Whoami(s.Client)
		if err != nil {
			return err
		}
		fmt.Fprintln(wOut, user)
	} else {
		fmt.Fprintf(wOut, "You are %s at %s\n", s.Username, s.Client.ControllerURL.String())
	}
	return nil
}

// Regenerate regenenerates a user's token.
func Regenerate(cf string, username string, all bool, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	token, err := auth.Regenerate(s.Client, username, all)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	if username == "" && !all {
		s.Client.Token = token
		_, err = s.Save(cf)

		if err != nil {
			return err
		}
	}

	fmt.Fprintln(wOut, "Token Regenerated")
	return nil
}

func readPassword() (string, error) {
	password, err := terminal.ReadPassword(int(syscall.Stdin))

	return string(password), err
}
