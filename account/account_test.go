package account

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/kiwih/nullables"
)

type DummyAccountStorer struct {
	OnlyAccount *Account
}

func (d DummyAccountStorer) LoadAccountFromEmail(email string) (*Account, error) {
	if d.OnlyAccount.Email == email {
		return d.OnlyAccount, nil
	}
	return nil, gorm.RecordNotFound
}

func (d DummyAccountStorer) LoadAccountFromId(id int64) (*Account, error) {
	return d.OnlyAccount, nil
}

func (d DummyAccountStorer) LoadAccountFromSession(sessionId string) (*Account, error) {
	return d.OnlyAccount, nil
}
func (d DummyAccountStorer) CreateAccount(a *Account) error {
	d.OnlyAccount = a
	return nil
}
func (d DummyAccountStorer) SaveAccount(a *Account) error {
	d.OnlyAccount = a
	return nil
}

var testStorage = DummyAccountStorer{
	OnlyAccount: &Account{
		Id:       1,
		Email:    "test@test",
		Password: "$2a$10$3NIEDlO7169hXn11bnIoGupnxlHmY7VB278/pxn4iIOFKqb8GGXaS", //bcrypt for "testing1+"
		Admin:    true,
		Nickname: "Test account",
		VoteBank: 10,
	},
}

func TestAttemptLogin(t *testing.T) {
	account, err := AttemptLogin(testStorage, "test@test", "not_testing1+", true)
	if err == nil {
		t.Fatal("Incorrect password logged in")
	} else if err != InvalidUsernameOrPassword {
		t.Fatal("Incorrect password not logged in but wrong error message: " + err.Error())
	}

	account, err = AttemptLogin(testStorage, "not_test@test", "testing1+", true)
	if err == nil {
		t.Fatal("Incorrect email logged in")
	} else if err != InvalidUsernameOrPassword {
		t.Fatal("Incorrect email not logged in but wrong error message: " + err.Error())
	}

	account, err = AttemptLogin(testStorage, "test@test", "testing1+", true)
	if err != nil {
		t.Fatal("Correct email/password did not log in")
	}

	if account.CurrentSession.Valid == false || len(account.CurrentSession.String) != 32 {
		t.Fatalf("Logging in did not set current session correctly: %+v", account.CurrentSession)
	}

	if account.SessionExpires.Valid == false {
		t.Fatalf("Logging in did not set session expires correctly: %+v", account.SessionExpires)
	}
}

func TestIsEmailInUse(t *testing.T) {
	inUse, _ := IsEmailInUse(testStorage, "not_test@test")
	if inUse {
		t.Fatal("Email incorrectly reported as being in use")
	}

	inUse, _ = IsEmailInUse(testStorage, "test@test")
	if !inUse {
		t.Fatal("Email incorrectly reported as not being in use")
	}
}

func TestIsPasswordAcceptable(t *testing.T) {
	testPasswords := map[string]bool{
		"qwerty":     false,
		"qwertyuiop": false,
		"asd12++":    false,
		"hjkYUO12":   true,
		"m1@7&":      false,
		"p99999999":  false,
		"rtyrty1+":   true,
	}

	for password, answer := range testPasswords {
		acceptable := IsPasswordAcceptable(password)
		if acceptable != answer {
			t.Fatalf("Password %s incorrectly reported as %v", password, acceptable)
		}
	}

}

func TestCanAccountBeMade(t *testing.T) {
	newAccount := &Account{
		Email:    "test@test",                         //currently, this should be taken
		Nickname: "Test account really long nickname", //this should be too long
		VoteBank: 10,
	}

	//don't need to do lots of passwords - these are tested by TestIsPasswordAcceptable
	if err := CanAccountBeMade(testStorage, newAccount, "rtyrty1+"); err == nil {
		t.Fatal("Account incorrectly able to be created")
	}

	newAccount.Nickname = "Test account"
	if err := CanAccountBeMade(testStorage, newAccount, "rtyrty1+"); err != EmailAddressAlreadyInUse {
		t.Fatal("Account incorrectly able to be created or incorrect error returned (email address should have been in use) - Actual error: ", err)
	}

	newAccount.Email = "something_not_in_use@test"
	if err := CanAccountBeMade(testStorage, newAccount, "rtyrty1+"); err != nil {
		t.Fatal("Account incorrectly not able to be created", err)
	}

	newAccount.Nickname = "Test account really long nickname"
	if err := CanAccountBeMade(testStorage, newAccount, "rtyrty1+"); err != AccountNicknameTooLong {
		t.Fatal("Account incorrectly able to be created or incorrect error returned (nickname should have been too long) - Actual error: ", err)
	}

}

func TestApplyVerificationCode(t *testing.T) {
	account := testStorage.OnlyAccount

	if err := account.ApplyVerificationCode(testStorage, "not_the_verification_code"); err != AccountDoesNotNeedVerification {
		t.Fatal("Verification code incorrectly able to be applied or incorrect error returned (account should not need verification) - Actual error: ", err)
	}
	var err error
	if account.VerificationCode, err = GenerateValidationKey(); err != nil {
		t.Fatal("Generating Validation Key failed:", err.Error())
	}

	if err := account.ApplyVerificationCode(testStorage, "not_the_verification_code"); err != AccountVerificationCodeNotMatch {
		t.Fatal("Verification code incorrectly able to be applied or incorrect error returned (verification code should not match) - Actual error: ", err)
	}

	if err := account.ApplyVerificationCode(testStorage, account.VerificationCode.String); err != nil {
		t.Fatal("Verification code incorrectly not be able to be applied - Actual error: ", err)
	}
}

func TestApplyPasswordResetVerificationCode(t *testing.T) {
	account := testStorage.OnlyAccount

	if err := account.ApplyPasswordResetVerificationCode(testStorage, "not_the_verification_code", "rtyrty1+"); err != AccountPasswordResetNotRequested {
		t.Fatal("Password Reset Verification code incorrectly able to be applied or incorrect error returned (account should not need password reset) - Actual error: ", err)
	}
	var err error
	if account.ResetPasswordVerificationCode, err = GenerateValidationKey(); err != nil {
		t.Fatal("Generating Validation Key failed:", err.Error())
	}

	if err := account.ApplyPasswordResetVerificationCode(testStorage, "not_the_verification_code", "rtyrty1+"); err != AccountVerificationCodeNotMatch {
		t.Fatal("Password Reset Verification code incorrectly able to be applied or incorrect error returned (verification code should not match) - Actual error: ", err)
	}

	if err := account.ApplyPasswordResetVerificationCode(testStorage, account.ResetPasswordVerificationCode.String, "rtyrty1+"); err != nil {
		t.Fatal("Password Reset Verification code incorrectly not be able to be applied - Actual error: ", err)
	}
}

func TestUpdateVoteBank(t *testing.T) {
	account := testStorage.OnlyAccount
	account.VoteBank = 10

	account.UpdateVoteBank(testStorage, true, 10)
	if account.VoteBank != 9 {
		t.Fatal("Casting a vote did not decrement vote bank correctly.")
	}

	account.UpdateVoteBank(testStorage, false, -10)
	if account.VoteBank != 8 {
		t.Fatal("Casting a vote did not decrement vote bank correctly.")
	}

	account.UpdateVoteBank(testStorage, true, -10)
	if account.VoteBank != 9 {
		t.Fatal("Refunding a vote did not increment vote bank correctly.")
	}

	account.UpdateVoteBank(testStorage, false, 10)
	if account.VoteBank != 10 {
		t.Fatal("Refunding a vote did not increment vote bank correctly.")
	}
}

func TestExpireSession(t *testing.T) {
	account := testStorage.OnlyAccount
	account.CurrentSession = nullables.NullString{String: "some_session_id", Valid: true}
	account.SessionExpires = nullables.NullTime{Time: time.Now().AddDate(0, 0, -1), Valid: true}

	account.ExpireSession(testStorage)

	time.Sleep(10 * time.Millisecond) //get some time to pass

	if account.CurrentSession.Valid == true || !account.SessionExpires.Time.Before(time.Now()) {
		t.Fatal("Session not expired correctly")
	}
}
