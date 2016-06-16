package fyidb

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	//_ "github.com/go-sql-driver/mysql"
	//	_ "github.com/denisenkom/go-mssqldb"

	"github.com/jinzhu/gorm"
	"github.com/kiwih/heyfyi/account"
	"github.com/kiwih/heyfyi/fact"
	"github.com/kiwih/nullables"
	"golang.org/x/crypto/bcrypt"
)

var dbGorm *gorm.DB

type DatabaseStorage struct {
	dbGorm *gorm.DB
}

var DbStorage DatabaseStorage

/* this is responsible for the creation of the connection to the database */
/* it routes the connection through GORM, the Go ORM manager */
func ConnectDatabase(dbname string) {
	//connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=ProfDev", *dbaddr, *dbuser, *dbpass, *dbport)
	//connString := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=True&loc=Local", dbuser, dbpass, dbaddr, dbport, dbname)
	connString := dbname + ".sqlite3"
	makeTables := false

	//check to see if the database doesn't exist. If it doesn't, we need to make the tables
	if _, err := os.Stat(connString); os.IsNotExist(err) {
		//mark tables for creation
		makeTables = true

	}

	dbConn, err := sql.Open("sqlite3", connString)
	if err != nil {
		log.Fatal("Could not open sqlite3 database connection: ", err.Error())
	}
	dbGormConnection, err := gorm.Open("sqlite3", dbConn)
	if err != nil {
		log.Fatal("Could not open GORM sqlite3 database connection: ", err.Error())
	}
	dbGormConnection.DB().SetMaxIdleConns(5)
	dbGormConnection.DB().SetMaxOpenConns(10)

	log.Println("Database connection opened.")

	dbGorm = dbGormConnection
	//dbGorm.LogMode(true)

	if makeTables == true {
		CreateDatabaseTables()
	} else {
		MigrateDatabaseTables()
	}

	DbStorage.dbGorm = dbGormConnection
}

func makeTable(tableName string, table interface{}) {
	log.Println("Making the " + tableName + " table...")
	if err := dbGorm.CreateTable(table).Error; err != nil {
		log.Println("An error occurred: " + err.Error())
	}
}

func migrateTable(tableName string, table interface{}) {
	log.Println("Migrating the " + tableName + " table...")
	if err := dbGorm.AutoMigrate(table).Error; err != nil {
		log.Println("An error occurred: " + err.Error())
	}
}

func MigrateDatabaseTables() {
	migrateTable("Accounts", &account.Account{})
	migrateTable("Facts", &fact.Fact{})
	migrateTable("References", &fact.Reference{})
	migrateTable("Votes", &fact.Vote{})
}

// this function is designed to be called to create the database tables when the appropriate flag is set
// if any tables preexist they will be dropped
func CreateDatabaseTables() {

	//we passed the test, let's create some tables
	makeTable("Accounts", &account.Account{})
	makeTable("Facts", &fact.Fact{})
	makeTable("References", &fact.Reference{})
	makeTable("Votes", &fact.Vote{})

	AddTestUser()
	AddTestFact()
}

func AddTestUser() {
	//add the test user to the database
	log.Printf("Inserting \"test@test\" as an Account...\n")
	hashedPass, err := bcrypt.GenerateFromPassword([]byte("testing1+"), 10)
	if err != nil {
		panic(err)
	}
	testUser := account.Account{
		Id:       1,
		Email:    "test@test",
		Password: string(hashedPass),
		Admin:    true,
		Nickname: "Test Admin",
		VoteBank: 100,
	}
	dbGorm.Create(&testUser)
}

func AddTestFact() {
	SpiderFact := fact.Fact{
		AccountId:      1,
		Fact:           "People almost never swallow spiders in their sleep",
		Explain:        "Yes! Humans vibrate while asleep - due to heartbeats, snoring, breathing, etc. Spiders treat vibrations as a danger signal, and so would quickly retreat from a slumbering body. In addition, beds don't feature any spider-friendly prey - so there is no motivation for the spider to come on to you.",
		ExplainFurther: "Spider experts concede that a sleeping person could swallow a spider, but it would be a strictly random and extremely unlikely event.",
		References: []fact.Reference{
			fact.Reference{
				Url:       "http://www.scientificamerican.com/article/fact-or-fiction-people-swallow-8-spiders-a-year-while-they-sleep1/",
				Publisher: "Scientific American",
				Title:     "Fact or Fiction? People Swallow 8 Spiders a Year While They Sleep",
			},
			fact.Reference{
				Url:       "http://www.cracked.com/article_16241_the-6-most-frequently-quoted-bullsh2At-statistics.html",
				Publisher: "Cracked",
				Title:     "The 6 Most Frequently Quoted Bull**** Statistics",
			},
		},
		Votes: []fact.Vote{
			fact.Vote{
				AccountId: 1,
				Score:     1,
			},
		},
	}

	dbGorm.Create(&SpiderFact)
}

func Special() {

}

//Loads an account given a username
func (s *DatabaseStorage) LoadAccountFromEmail(email string) (*account.Account, error) {
	var a account.Account
	if err := s.dbGorm.Where(&account.Account{Email: email}).Find(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

//Loads an account given an ID
func (s *DatabaseStorage) LoadAccountFromId(id int64) (*account.Account, error) {
	var a account.Account
	if err := s.dbGorm.Find(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *DatabaseStorage) LoadAccountFromSession(sessionId string) (*account.Account, error) {
	var a account.Account
	if err := s.dbGorm.Where(&account.Account{CurrentSession: nullables.NullString{String: sessionId, Valid: true}}).Find(&a).Error; err != nil {
		return nil, err
	}
	if a.SessionExpires.Valid { //allow users to have infinite sessions if the SessionExpires is null, but this is an edge case
		if a.SessionExpires.Time.Before(time.Now()) {
			return nil, errors.New("Session expired.")
		}
	}
	return &a, nil

}

func (s *DatabaseStorage) CreateAccount(a *account.Account) error {
	return s.dbGorm.Create(a).Error
}

func (s *DatabaseStorage) SaveAccount(a *account.Account) error {
	return s.dbGorm.Save(a).Error
}

func (s *DatabaseStorage) LoadFactFromId(id int64) (*fact.Fact, error) {
	var f fact.Fact
	if err := s.dbGorm.Find(&f, id).Related(&f.References).Related(&f.Votes).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *DatabaseStorage) CreateFact(f *fact.Fact) error {
	f.AwaitModeration = true
	return s.dbGorm.Create(f).Error
}

func (s *DatabaseStorage) DeleteFact(f *fact.Fact) error {
	return s.dbGorm.Delete(&f).Error
}

//Record a vote against a fact. Return the vote, and any applicable error.
func (s *DatabaseStorage) GetVoteForFact(accountId int64, factId int64) (*fact.Vote, error) {
	var v fact.Vote
	if err := s.dbGorm.FirstOrCreate(&v, fact.Vote{AccountId: accountId, FactId: factId}).Error; err != nil {
		return nil, err
	}

	return &v, nil
}

func (s *DatabaseStorage) SaveVote(v *fact.Vote) error {
	return s.dbGorm.Save(&v).Error
}

func (s *DatabaseStorage) ModerateFact(f *fact.Fact, enable bool) error {
	f.AwaitModeration = enable
	return s.dbGorm.Save(f).Error
}

func (s *DatabaseStorage) ListFacts(accountId int64, viewUnmoderated bool) ([]fact.Fact, error) {
	var facts []fact.Fact

	if viewUnmoderated {
		if err := s.dbGorm.Find(&facts).Error; err != nil {
			return nil, err
		}
	} else {
		//if not viewing unmoderated, only show facts that are awaiting moderation that are yours
		if accountId > 0 {
			if err := s.dbGorm.Where("facts.await_moderation = 0 or facts.account_id = ?", accountId).Find(&facts).Error; err != nil {
				return nil, err
			}
		} else {
			if err := s.dbGorm.Where("facts.await_moderation = 0").Find(&facts).Error; err != nil {
				return nil, err
			}
		}
	}

	return facts, nil
}

func (s *DatabaseStorage) GiveOneVoteToAllAccounts() error {
	log.Println("Giving vote to all accounts.")
	return s.dbGorm.Model(&account.Account{}).UpdateColumn("vote_bank", gorm.Expr("vote_bank + ?", 1)).Error
}
