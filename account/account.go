package account

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	"unicode"

	"github.com/jinzhu/gorm"
	"github.com/kiwih/nullables"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
)

type Account struct {
	Id                            int64
	Email                         string               `sql:"unique; type:varchar(60);" validate:"nonzero"`
	Nickname                      string               `sql:"type:varchar(15); validate:"nonzero"`
	Password                      string               `sql:"type:varchar(60);"`
	VerificationCode              nullables.NullString `sql:"type:varchar(32)"`
	ResetPasswordVerificationCode nullables.NullString `sql:"type:varchar(32)"`
	CurrentSession                nullables.NullString `sql:"type:varchar(32)"`
	SessionExpires                nullables.NullTime
	VoteBank                      int64
	Admin                         bool
	CreatedAt                     nullables.NullTime
	UpdatedAt                     nullables.NullTime
	DeletedAt                     nullables.NullTime
}

type AccountStorer interface {
	LoadAccountFromEmail(email string) (*Account, error)
	LoadAccountFromId(int64) (*Account, error)
	LoadAccountFromSession(sessionId string) (*Account, error)
	CreateAccount(*Account) error
	SaveAccount(*Account) error
}

var (
	NoVotesLeft                      error = errors.New("No votes left in bank!")
	InvalidUsernameOrPassword        error = errors.New("Invalid Username/Password")
	AccountNotYetVerified            error = errors.New("Account not yet verified.")
	AccountNicknameTooLong           error = errors.New("Your nickname cannot be longer than 15 characters!")
	AccountDoesNotNeedVerification   error = errors.New("Account doesn't need verifying!")
	AccountVerificationCodeNotMatch  error = errors.New("Bad verification code!")
	AccountPasswordResetNotRequested error = errors.New("Password reset not requested!")
	EmailAddressAlreadyInUse         error = errors.New("This email address is already in use!")
	PasswordNotAcceptable            error = errors.New("Password must contain at least 3 of types of characters from uppercase, lowercase, punctuation, and digits, and be at least 8 characters long.")
)

func GenerateValidationKey() (nullables.NullString, error) {
	//generate validation key
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return nullables.NullString{}, err
	}

	return nullables.NullString{String: fmt.Sprintf("%x", md5.Sum(b)), Valid: true}, nil
}

func AttemptLogin(as AccountStorer, propEmail string, propPassword string, willExpire bool) (*Account, error) {
	propUser, err := as.LoadAccountFromEmail(propEmail)

	if err == nil {
		if err = bcrypt.CompareHashAndPassword([]byte(propUser.Password), []byte(propPassword)); err == nil {
			//they have passed the login check. Save them to the session and redirect to management portal
			if propUser.VerificationCode.Valid {
				return nil, AccountNotYetVerified
			}
			//successful login.
			//generate session
			propUser.CurrentSession, err = GenerateValidationKey()
			if err != nil {
				return nil, err
			}
			if willExpire {
				propUser.SessionExpires = nullables.NullTime{Time: time.Now().Add(3600 * time.Second), Valid: true} //expires in 1 hour if they don't want to be remembered
			} else {
				propUser.SessionExpires = nullables.NullTime{Time: time.Now().AddDate(0, 1, 0), Valid: true} //expires in 1 month if they want to be remembered
			}
			return propUser, as.SaveAccount(propUser)
		}
		return nil, InvalidUsernameOrPassword
	}
	return nil, InvalidUsernameOrPassword
}

func IsEmailInUse(as AccountStorer, email string) (bool, error) {
	if _, err := as.LoadAccountFromEmail(email); err != nil {
		if err.Error() == gorm.RecordNotFound.Error() {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

var passwordMustHave = []func(rune) bool{
	unicode.IsUpper,
	unicode.IsLower,
	unicode.IsSymbol,
	unicode.IsDigit,
	unicode.IsPunct,
}

func IsPasswordAcceptable(plaintextPassword string) bool {
	if len(plaintextPassword) < 8 {
		return false
	}

	types := 0

	for _, testRune := range passwordMustHave {
		found := false
		for _, r := range plaintextPassword {
			if testRune(r) {
				found = true
			}
		}
		if found {
			types++
		}
	}

	return types >= 3
}

func CanAccountBeMade(as AccountStorer, a *Account, passwordClearText string) error {
	//make sure password is secure
	if !IsPasswordAcceptable(passwordClearText) {
		return PasswordNotAcceptable
	}
	//make sure username isn't taken
	if nameTaken, err := IsEmailInUse(as, a.Email); nameTaken == true || err != nil {
		if err != nil {
			return err
		}
		return EmailAddressAlreadyInUse
	}

	if len(a.Nickname) > 15 {
		return AccountNicknameTooLong
	}

	//check account is valid
	if err := a.Valid(); err != nil {
		return err
	}
	return nil

}

func (a *Account) SetPassword(cleartext string) error {
	hashpass, err := bcrypt.GenerateFromPassword([]byte(cleartext), 10)
	if err != nil {
		return err
	}
	a.Password = string(hashpass)
	return nil
}

func CheckAndCreateAccount(as AccountStorer, email string, password string, nickname string) error {
	//check if account is valid

	a := &Account{
		Nickname: nickname,
		Email:    email,
		VoteBank: 10,
	}

	if err := CanAccountBeMade(as, a, password); err != nil {
		return err
	}

	//hash password
	if err := a.SetPassword(password); err != nil {
		log.Println("Error in password hashing: " + err.Error())
		return err
	}

	var err error
	if a.VerificationCode, err = GenerateValidationKey(); err != nil {
		return err
	}

	//create account, send validation email
	if err = as.CreateAccount(a); err != nil {
		return err
	}

	sendEmail(a.Email, "Verification code", "Hello!\r\n\r\nTo validate your hey.fyi account, you need to follow this link:\r\nhttp://hey.fyi/verify/"+strconv.FormatInt(a.Id, 10)+"/"+a.VerificationCode.String+"\r\n\r\nI hope you enjoy using the service!\r\n\r\nRegards,\r\nhey.fyi")
	log.Printf("Verification code for user %s is %s\n", a.Email, a.VerificationCode.String)

	return nil
}

func DoPasswordResetRequestIfPossible(as AccountStorer, email string) error {
	a, err := as.LoadAccountFromEmail(email)
	if err != nil {
		return err
	}

	if a.ResetPasswordVerificationCode, err = GenerateValidationKey(); err != nil {
		return err
	}

	sendEmail(a.Email, "Password Reset Request", "Hello!\r\n\r\nSomeone requested a password reset to your hey.fyi account.\r\nIf you didn't request this, simply ignore this email.\r\n\r\nOtherwise, follow this link:\r\nhttp://hey.fyi/reset/"+strconv.FormatInt(a.Id, 10)+"/"+a.ResetPasswordVerificationCode.String+"\r\n\r\nRegards,\r\nhey.fyi")
	log.Printf("Reset Password verification code for user %s is %s\n", a.Email, a.VerificationCode.String)

	return as.SaveAccount(a)
}

func (a Account) Valid() error {
	return validator.Validate(a)
}

func (a *Account) ApplyVerificationCode(as AccountStorer, verificationCode string) error {
	if a.VerificationCode.Valid == false {
		return AccountDoesNotNeedVerification
	}

	if a.VerificationCode.String != verificationCode {
		return AccountVerificationCodeNotMatch
	}

	a.VerificationCode.String = ""
	a.VerificationCode.Valid = false

	return as.SaveAccount(a)
}

func (a *Account) AwaitingPasswordReset() bool {
	return a.ResetPasswordVerificationCode.Valid
}

func (a *Account) ApplyPasswordResetVerificationCode(as AccountStorer, resetVerificationCode string, newPassword string) error {
	if a.AwaitingPasswordReset() == false {
		return AccountPasswordResetNotRequested
	}

	if a.ResetPasswordVerificationCode.String != resetVerificationCode {
		return AccountVerificationCodeNotMatch
	}

	if !IsPasswordAcceptable(newPassword) {
		return PasswordNotAcceptable
	}

	if err := a.SetPassword(newPassword); err != nil {
		return err
	}

	a.ResetPasswordVerificationCode.String = ""
	a.ResetPasswordVerificationCode.Valid = false

	return as.SaveAccount(a)
}

func (a *Account) UpdateVoteBank(as AccountStorer, up bool, currentAccountVote int64) error {
	if (up && currentAccountVote >= 0) || (!up && currentAccountVote <= 0) { //they are casting a vote
		if a.VoteBank <= 0 {
			return NoVotesLeft
		} else {
			a.VoteBank--
			return as.SaveAccount(a)
		}
	}
	//else, they are retracting a vote
	a.VoteBank++
	return as.SaveAccount(a)
}

func (a *Account) ExpireSession(as AccountStorer) error {
	a.CurrentSession.Valid = false
	a.CurrentSession.String = ""

	a.SessionExpires.Time = time.Now()

	return as.SaveAccount(a)
}
