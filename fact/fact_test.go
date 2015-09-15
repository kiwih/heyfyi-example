package fact

import (
	"testing"

	"github.com/jinzhu/gorm"
)

type DummyFactStorer struct {
	OnlyFact *Fact
}

func (d DummyFactStorer) ListFacts(accountId int64, awaitModeration bool) ([]Fact, error) {
	f := make([]Fact, 1)
	f[0] = *d.OnlyFact
	return f, nil
}

func (d DummyFactStorer) LoadFactFromId(id int64) (*Fact, error) {
	return d.OnlyFact, nil
}

func (d DummyFactStorer) CreateFact(f *Fact) error {
	d.OnlyFact = f
	return nil
}

func (d DummyFactStorer) SaveFact(f *Fact) error {
	d.OnlyFact = f
	return nil
}

func (d DummyFactStorer) DeleteFact(f *Fact) error {
	return nil
}

func (d DummyFactStorer) GetVoteForFact(accountId int64, factId int64) (*Vote, error) {
	for i := 0; i < len(d.OnlyFact.Votes); i++ {
		if d.OnlyFact.Votes[i].AccountId == accountId {
			return &d.OnlyFact.Votes[i], nil
		}
	}

	d.OnlyFact.Votes = append(d.OnlyFact.Votes, Vote{
		Id:        int64(len(d.OnlyFact.Votes) + 1),
		AccountId: accountId,
		FactId:    factId,
		Score:     0,
	})

	return &d.OnlyFact.Votes[len(d.OnlyFact.Votes)-1], nil
}

func (d DummyFactStorer) SaveVote(v *Vote) error {
	for i := 0; i < len(d.OnlyFact.Votes); i++ {
		if d.OnlyFact.Votes[i].Id == v.Id {
			d.OnlyFact.Votes[i] = *v
			return nil
		}
	}
	return gorm.RecordNotFound
}

func (d DummyFactStorer) ModerateFact(f *Fact, enable bool) error {
	f.AwaitModeration = enable
	return nil
}

var testStorage = DummyFactStorer{
	OnlyFact: nil,
}

var testFact = Fact{
	Id:              1,
	Fact:            "This is the fact string",
	Explain:         "This is the explain string",
	ExplainFurther:  "This is the explain further string",
	AwaitModeration: false,
	References: []Reference{
		Reference{
			Id:        1,
			FactId:    1,
			Url:       "http://example.com",
			Publisher: "example publisher",
			Title:     "example title",
		},
	},
	Votes: []Vote{
		Vote{
			Id:        1,
			FactId:    1,
			AccountId: 1,
			Score:     1,
		},
	},
	AccountId: 1,
}

func TestCreateFactAndValidateReferences(t *testing.T) {
	tempFact := testFact //i don't want to edit testFact
	tempFact.Fact = ""

	if err := CreateFact(testStorage, &tempFact); err != AllFieldsAreCompulsory {
		t.Fatal("Did not return AllFieldsAreCompulsor when missing Fact field")
	}

	tempFact.Fact = "This is the fact string"
	tempFact.Explain = ""

	if err := CreateFact(testStorage, &tempFact); err != AllFieldsAreCompulsory {
		t.Fatal("Did not return AllFieldsAreCompulsor when missing Explain field")
	}

	tempFact.Explain = "This is the explain string"
	tempFact.ExplainFurther = ""

	if err := CreateFact(testStorage, &tempFact); err != AllFieldsAreCompulsory {
		t.Fatal("Did not return AllFieldsAreCompulsor when missing Explain Further")
	}

	tempFact.ExplainFurther = "This is the explainfurther string"

	if err := CreateFact(testStorage, &tempFact); err != NotEnoughReferences {
		t.Fatal("NotEnoughReferences was not thrown when only one reference provided")
	}

	//we can just use the ValidateReferences function too, as it is called by CreateFact
	tempFact.References = append(tempFact.References, Reference{})
	if err := tempFact.ValidateReferences(); err != AllFieldsAreCompulsory {
		t.Fatal("AllFieldsAreCompulsory was not thrown when empty fields (all) present")
	}

	tempFact.References[1].Title = "Some title"
	if err := tempFact.ValidateReferences(); err != AllFieldsAreCompulsory {
		t.Fatal("AllFieldsAreCompulsory was not thrown when empty fields (all except title) present")
	}

	tempFact.References[1].Publisher = "Some publisher"
	if err := tempFact.ValidateReferences(); err != AllFieldsAreCompulsory {
		t.Fatal("AllFieldsAreCompulsory was not thrown when empty fields (just url) present")
	}

	tempFact.References[1].Url = "example.com"
	err := tempFact.ValidateReferences()
	if err != nil {
		t.Fatal("An error was thrown when it shouldn't have: " + err.Error())
	}
	if tempFact.References[1].Url != "http://example.com" {
		t.Fatal("http:// not added to URL")
	}

	tempFact.References[1].Url = "http://example.com"
	err = tempFact.ValidateReferences()
	if err != nil {
		t.Fatal("An error was thrown when it shouldn't have: " + err.Error())
	}
	if tempFact.References[1].Url != "http://example.com" {
		t.Fatal("URL changed when it shouldn't have been: " + tempFact.References[1].Url)
	}

	tempFact.References[1].Url = "ftp://example.com"
	err = tempFact.ValidateReferences()
	if err != nil {
		t.Fatal("An error was thrown when it shouldn't have: " + err.Error())
	}
	if tempFact.References[1].Url != "ftp://example.com" {
		t.Fatal("URL changed when it shouldn't have been: " + tempFact.References[1].Url)
	}

	tempFact.References[1].Url = "https://example.com"
	err = tempFact.ValidateReferences()
	if err != nil {
		t.Fatal("An error was thrown when it shouldn't have: " + err.Error())
	}
	if tempFact.References[1].Url != "https://example.com" {
		t.Fatal("URL changed when it shouldn't have been: " + tempFact.References[1].Url)
	}

	tempFact.AccountId = 0
	if err := CreateFact(testStorage, &tempFact); err != NoAccountSpecified {
		t.Fatal("NoAccountSpecfied was not thrown when no AccountID provided")
	}

	tempFact.AccountId = 1
	if err := CreateFact(testStorage, &tempFact); err != nil {
		t.Fatal("Error thrown when it shouldn't have been: " + err.Error())
	}

}

func TestVoteForFactAndGetScore(t *testing.T) {
	tempFact := testFact //i don't want to edit testFact
	testStorage.OnlyFact = &tempFact

	testStorage.OnlyFact.Votes[0].Score = 1

	if v, err := VoteForFact(testStorage, 1, 1, true); v.Score != 2 || v.AccountId != 1 || err != nil {
		t.Fatalf("VoteForFact did not work, vote did not get changed correctly, output %+v\n", v)
	}

	if v, err := VoteForFact(testStorage, 2, 1, true); v.Score != 1 || v.AccountId != 2 || err != nil {
		t.Fatalf("VoteForFact did not work, vote did not get made correctly, output %+v\n", v)
	}

	if v, err := VoteForFact(testStorage, 3, 1, false); v.Score != -1 || v.AccountId != 3 || err != nil {
		t.Fatalf("VoteForFact did not work, vote did not get made correctly, output %+v\n", v)
	}

	if v, err := VoteForFact(testStorage, 1, 1, false); v.Score != 1 || v.AccountId != 1 || err != nil {
		t.Fatalf("VoteForFact did not work, vote did not get changed correctly, output %+v\n", v)
	}

	if score := testStorage.OnlyFact.GetScore(1); score.Ups != 2 || score.Downs != 1 || score.AccountVote != 1 {
		t.Fatalf("GetScore did not work, vote did not get changed correctly, output %+v\n", score)
	}
}

func TestValidateReferences(t *testing.T) {

}
