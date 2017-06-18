package fact

import (
	"errors"
	"strings"

	"github.com/kiwih/nullables"
)

type Fact struct {
	Id              int64
	Fact            string
	Explain         string
	ExplainFurther  string
	AwaitModeration bool
	References      []Reference
	Votes           []Vote
	AccountId       int64
	CreatedAt       nullables.NullTime
	EditedAt        nullables.NullTime
	DeletedAt       nullables.NullTime
}

type Reference struct {
	Id        int64
	FactId    int64
	Url       string
	Publisher string
	Title     string
	CreatedAt nullables.NullTime
	EditedAt  nullables.NullTime
	DeletedAt nullables.NullTime
}

type Vote struct {
	Id        int64
	FactId    int64
	AccountId int64
	Score     int64
	CreatedAt nullables.NullTime
	DeletedAt nullables.NullTime
}

type FactStorer interface {
	ListFacts(accountId int64, awaitModeration bool) ([]Fact, error)
	LoadFactFromId(id int64) (*Fact, error)
	DeleteFact(*Fact) error
	CreateFact(*Fact) error
	GetVoteForFact(accountId int64, factId int64) (*Vote, error)
	SaveVote(*Vote) error
	ModerateFact(f *Fact, enable bool) error
}

type VoteScore struct {
	Ups         int64
	Downs       int64
	AccountVote int64
}

var (
	NotEnoughReferences    = errors.New("You need at least 2 references!")
	NoAccountSpecified     = errors.New("No account ID was specified!")
	AllFieldsAreCompulsory = errors.New("All fields are compulsory.")
)

func VoteForFact(fs FactStorer, accountId int64, factId int64, up bool) (*Vote, error) {
	var scoreChange int64 = 0
	if up {
		scoreChange = 1
	} else {
		scoreChange = -1
	}

	vote, err := fs.GetVoteForFact(accountId, factId)
	if err != nil {
		return nil, err
	}

	vote.Score += scoreChange

	return vote, fs.SaveVote(vote)

}

func CreateFact(fs FactStorer, f *Fact) error {
	if f.Fact == "" || f.Explain == "" || f.ExplainFurther == "" {
		return AllFieldsAreCompulsory
	}

	if err := f.ValidateReferences(); err != nil {
		return err
	}

	if f.AccountId == 0 {
		return NoAccountSpecified
	}

	return fs.CreateFact(f)
}

func (f *Fact) GetScore(currentAccountId int64) VoteScore {
	var v VoteScore

	for _, vote := range f.Votes {
		if vote.AccountId == currentAccountId {
			v.AccountVote = vote.Score
		}
		if vote.Score > 0 {
			v.Ups += vote.Score
		} else {
			v.Downs -= vote.Score
		}
	}
	return v
}

//if the references are fine, no error will be returned
//this will also add "http://" to the urls if they do not begin with it or https:// or ftp://
func (f *Fact) ValidateReferences() error {
	if len(f.References) < 2 {
		return NotEnoughReferences
	}

	for i := 0; i < len(f.References); i++ {
		if f.References[i].Url == "" || f.References[i].Title == "" || f.References[i].Publisher == "" {
			return AllFieldsAreCompulsory
		}
		if strings.Contains(f.References[i].Url, "https://") {
			continue
		}
		if strings.Contains(f.References[i].Url, "http://") {
			continue
		}
		if strings.Contains(f.References[i].Url, "ftp://") {
			continue
		}
		f.References[i].Url = "http://" + f.References[i].Url
	}
	return nil
}
