package instagram

import (
	"fmt"
	"goinstabot/config"
	"goinstabot/db"
	"log"
	"strings"
	"time"

	goinsta "github.com/ahmdrz/goinsta"
)

var (
	Insta          *goinsta.Instagram
	Following      = make(map[string]int)
	Blocked        = make(map[string]int)
	PreferredNames = make(map[string]int)
	FollowingList  []string
)

func Login() {
	Insta = goinsta.New(config.Localconfig.InstaUser, config.Localconfig.InstaPass)
	if err := Insta.Login(); err != nil {
		fmt.Println(err)
		return
	}
}
func LoadFollowingFromDB() {
	followingusersDB := getAllFollowing_FromDB()
	FollowingList = followingusersDB
}
func StartFollowingWithMediaLikes(Limit int) {
	for _, myUser := range FollowingList {

		user, err := Insta.Profiles.ByName(myUser)
		log.Println("Checking user: " + myUser)

		if err != nil {
			log.Println(err)
		}
		media := user.Feed()
		for media.Next() {
			fmt.Printf("Printing %d items\n", len(media.Items))
			for _, item := range media.Items {
				time.Sleep(1 * time.Second)
				for _, liker := range item.Likers {
					fullname := strings.Split(liker.FullName, " ")
					firstname := strings.ToLower(fullname[0])
					if PreferredNames[firstname] == 1 && Blocked[liker.Username] != 1 && Following[liker.Username] != 1 {
						profile, err := Insta.Profiles.ByID(liker.ID)
						if err != nil {
							continue
						}
					}
				}
			}
			break
		}
	}
}

func SyncMappings(followingList []string, blockedList []string) {
	Following = make(map[string]int)
	for _, follow := range followingList {
		Following[follow] = 1
	}
	Blocked = make(map[string]int)
	for _, block := range blockedList {
		Blocked[block] = 1
	}
}
func SyncFollowingDBfromApp() {
	var followingusersDB []string
	var followingusersApp []string
	followingusersDB = getAllFollowing_FromDB()
	followingusersApp = getAllFollowing_FromInstagram()
	FollowingList = followingusersApp
	if len(followingusersApp) > 0 {
	DBvsApp:
		for _, dbu := range followingusersDB {
			var found bool = false
			for _, dbapp := range followingusersApp {
				if dbu == dbapp {
					found = true
				}
			}
			if !found {
				dberr := db.DBDeletePostgres_Following(dbu)
				if dberr != nil {
					continue DBvsApp
				}
				dberr = db.DBInsertPostgres_Blocked(dbu)
				if dberr != nil {
					continue DBvsApp
				}
				// This means the user was there before in DB but was deleted from App (unfollowed manually)
				// TODO: Logic to Block it in instagram

			}
		}

	AppvsDB:
		for _, dbapp := range followingusersApp {
			var found bool = false
			for _, dbu := range followingusersDB {
				if dbu == dbapp {
					found = true
				}
			}
			if !found {
				// This means the user was added and DB doesnt have it , so we should add it there
				dberr := db.DBInsertPostgres_Following(dbapp)
				if dberr != nil {
					continue AppvsDB
				}
			}
		}

	}
	SyncMappings(followingusersApp, db.DBSelectPostgres_Blocked())

}

func getAllFollowing_FromDB() []string {
	var followingusers []string
	followingusers = db.DBSelectPostgres_Following()
	return followingusers
}
func getAllFollowing_FromInstagram() []string {
	var followingusers []string
	users := Insta.Account.Following()
	for users.Next() {
		for _, user := range users.Users {
			followingusers = append(followingusers, user.Username)
		}
	}
	return followingusers
}