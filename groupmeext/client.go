package groupmeext

import (
	"context"

	"github.com/beeper/groupme-lib"
	log "maunium.net/go/maulogger/v2"
)

type Client struct {
	*groupme.Client
	log log.Logger
}

// NewClient creates a new GroupMe API Client
func NewClient(authToken string, log log.Logger) *Client {
	n := Client{
		Client: groupme.NewClient(authToken),
		log:    log,
	}
	return &n
}
func (c Client) IndexAllGroups() ([]*groupme.Group, error) {
	groups, err := c.IndexGroups(context.TODO(), &groupme.GroupsQuery{
		//	Omit:    "memberships",
		PerPage: 100, //TODO: Configurable and add multipage support
	})
	if err != nil {
		c.log.Warnln("Failed to index groups:", err)
	}
	return groups, err
}

func (c Client) IndexAllRelations() ([]*groupme.User, error) {
	users, err := c.IndexRelations(context.TODO())
	if err != nil {
		c.log.Warnln("Failed to index relations:", err)
	}
	return users, err
}

func (c Client) IndexAllChats() ([]*groupme.Chat, error) {
	chats, err := c.IndexChats(context.TODO(), &groupme.IndexChatsQuery{
		PerPage: 100, //TODO?
	})
	if err != nil {
		c.log.Warnln("Failed to index chats:", err)
	}
	return chats, err
}

func (c Client) LoadMessagesAfter(groupID groupme.ID, lastMessageID string, lastMessageFromMe bool, private bool) ([]*groupme.Message, error) {
	if private {
		ans, e := c.IndexDirectMessages(context.TODO(), groupID.String(), &groupme.IndexDirectMessagesQuery{
			SinceID: groupme.ID(lastMessageID),
			//Limit:    num,
		})
		//fmt.Println(groupID, lastMessageID, num, i.Count, e)
		if e != nil {
			return nil, e
		}

		for i, j := 0, len(ans.Messages)-1; i < j; i, j = i+1, j-1 {
			ans.Messages[i], ans.Messages[j] = ans.Messages[j], ans.Messages[i]
		}
		return ans.Messages, nil
	} else {
		i, e := c.IndexMessages(context.TODO(), groupID, &groupme.IndexMessagesQuery{
			AfterID: groupme.ID(lastMessageID),
			//20 for consistency with dms
			Limit: 20,
		})
		//fmt.Println(groupID, lastMessageID, num, i.Count, e)
		if e != nil {
			return nil, e
		}
		return i.Messages, nil
	}
}

func (c Client) LoadMessagesBefore(groupID, lastMessageID string, private bool) ([]*groupme.Message, error) {
	if private {
		i, e := c.IndexDirectMessages(context.TODO(), groupID, &groupme.IndexDirectMessagesQuery{
			BeforeID: groupme.ID(lastMessageID),
			//Limit:    num,
		})
		//fmt.Println(groupID, lastMessageID, num, i.Count, e)
		if e != nil {
			return nil, e
		}
		return i.Messages, nil
	} else {
		//TODO: limit max 100
		i, e := c.IndexMessages(context.TODO(), groupme.ID(groupID), &groupme.IndexMessagesQuery{
			BeforeID: groupme.ID(lastMessageID),
			//20 for consistency with dms
			Limit: 20,
		})
		//fmt.Println(groupID, lastMessageID, num, i.Count, e)
		if e != nil {
			return nil, e
		}
		return i.Messages, nil
	}
}

func (c *Client) RemoveFromGroup(uid, groupID groupme.ID) error {
	group, err := c.ShowGroup(context.TODO(), groupID)
	if err != nil {
		return err
	}
	return c.RemoveMember(context.TODO(), groupID, group.GetMemberByUserID(uid).ID)
}
