package main

import (
	"testing"
	"time"

	"github.com/beeper/groupme-lib"
	"github.com/beeper/groupme/config"
	"github.com/beeper/groupme/database"
	_ "github.com/mattn/go-sqlite3"
	"maunium.net/go/maulogger/v2"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/util/dbutil"
)

func newTestUser(t *testing.T) (*GMBridge, *User) {
	db, err := dbutil.NewFromConfig("sqlite3", ":memory:", dbutil.MauLogger(maulogger.Create()))
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
    CREATE TABLE user (
        gmid TEXT PRIMARY KEY,
        mxid TEXT,
        auth_token TEXT,
        management_room TEXT,
        space_room TEXT
    );
    CREATE TABLE portal (
        gmid TEXT,
        receiver TEXT,
        mxid TEXT,
        name TEXT,
        name_set BOOLEAN,
        topic TEXT,
        topic_set BOOLEAN,
        avatar TEXT,
        avatar_url TEXT,
        avatar_set BOOLEAN,
        encrypted BOOLEAN,
        PRIMARY KEY (gmid, receiver)
    );
    CREATE TABLE user_portal (
        user_mxid TEXT,
        portal_gmid TEXT,
        portal_receiver TEXT,
        in_space BOOLEAN,
        PRIMARY KEY (user_mxid, portal_gmid, portal_receiver)
    );
    `)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	br := &GMBridge{
		DB:     database.New(db.Database),
		Config: &config.Config{},
		Log:    maulogger.Create(),
	}
	br.Config.Bridge.PersonalFilteringSpaces = true

	user := br.NewUser(br.DB.User.New())
	user.MXID = id.NewUserID("testuser", "example.com")
	user.GMID = "123456"
	user.inSpaceCache = make(map[database.PortalKey]bool)
	return br, user
}

func TestSyncPortals_NewGroup(t *testing.T) {
	br, user := newTestUser(t)

	// Mock GroupList
	groupID := "11111"
	user.GroupList = map[groupme.ID]groupme.Group{
		groupme.ID(groupID): {
			ID:        groupme.ID(groupID),
			Name:      "Test Group",
			UpdatedAt: groupme.Time{Time: time.Now()},
		},
	}

	// We need to ensure GetPortalByGMID returns something or we create it?
	// In syncPortals: portal := user.bridge.GetPortalByGMID(database.GroupPortalKey(group.ID))
	// GetPortalByGMID checks DB.
	// If it's nil, we might have issues if Sync interacts with it.
	// user.GetPortalByGMID creates a New() portal if not found in DB?
	// Let's check `database/portal.go`: GetByGMID returns *Portal.
	// `GMBridge.GetPortalByGMID` calls `br.loadDBPortal`.
	// We need to make sure `br.GetPortalByGMID` behaves correctly.
	// Ideally we modify GMBridge to have portalsByGMID map initialized.
	br.portalsByGMID = make(map[database.PortalKey]*Portal)
	br.portalsByMXID = make(map[id.RoomID]*Portal)

	user.syncPortals(false)

	// Check if portal was added to user_portal table (marked in space)
	portalKey := database.GroupPortalKey(groupme.ID(groupID))
	if !user.IsInSpace(portalKey) {
		t.Errorf("Expected portal %s to be in space", groupID)
	}

	// Verify SetPortalKeys logic (effectively verified by IsInSpace check above)
}

func TestSyncPortals_ExistingGroup(t *testing.T) {
	br, user := newTestUser(t)

	groupID := "22222"
	portalKey := database.GroupPortalKey(groupme.ID(groupID))

	// Pre-create portal in DB
	dbPortal := br.DB.Portal.New()
	dbPortal.Key = portalKey
	dbPortal.Insert()

	user.GroupList = map[groupme.ID]groupme.Group{
		groupme.ID(groupID): {
			ID:        groupme.ID(groupID),
			Name:      "Existing Group",
			UpdatedAt: groupme.Time{Time: time.Now()},
		},
	}

	br.portalsByGMID = make(map[database.PortalKey]*Portal)
	br.portalsByMXID = make(map[id.RoomID]*Portal)

	user.syncPortals(false)

	if !user.IsInSpace(portalKey) {
		t.Errorf("Expected existing portal %s to be in space", groupID)
	}
}
