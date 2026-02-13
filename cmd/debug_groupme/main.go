package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/beeper/groupme-lib"
	"github.com/beeper/groupme/groupmeext"
	"maunium.net/go/maulogger/v2"
)

// SimpleLogger satisfies the maulogger.Logger interface
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(args ...interface{}) {
	fmt.Println(append([]interface{}{"[DEBUG]"}, args...)...)
}
func (l *SimpleLogger) Debugf(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}
func (l *SimpleLogger) Debugln(args ...interface{}) {
	fmt.Println(append([]interface{}{"[DEBUG]"}, args...)...)
}

// For Debugfln support if needed (sometimes people alias it)
func (l *SimpleLogger) Debugfln(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}

func (l *SimpleLogger) Info(args ...interface{}) {
	fmt.Println(append([]interface{}{"[INFO]"}, args...)...)
}
func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}
func (l *SimpleLogger) Infoln(args ...interface{}) {
	fmt.Println(append([]interface{}{"[INFO]"}, args...)...)
}
func (l *SimpleLogger) Infofln(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (l *SimpleLogger) Warn(args ...interface{}) {
	fmt.Println(append([]interface{}{"[WARN]"}, args...)...)
}
func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}
func (l *SimpleLogger) Warnln(args ...interface{}) {
	fmt.Println(append([]interface{}{"[WARN]"}, args...)...)
}
func (l *SimpleLogger) Warnfln(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}

func (l *SimpleLogger) Error(args ...interface{}) {
	fmt.Println(append([]interface{}{"[ERROR]"}, args...)...)
}
func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}
func (l *SimpleLogger) Errorln(args ...interface{}) {
	fmt.Println(append([]interface{}{"[ERROR]"}, args...)...)
}
func (l *SimpleLogger) Errorfln(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

func (l *SimpleLogger) Fatal(args ...interface{}) {
	fmt.Println(append([]interface{}{"[FATAL]"}, args...)...)
	os.Exit(1)
}
func (l *SimpleLogger) Fatalf(format string, args ...interface{}) {
	fmt.Printf("[FATAL] "+format+"\n", args...)
	os.Exit(1)
}
func (l *SimpleLogger) Fatalln(args ...interface{}) {
	fmt.Println(append([]interface{}{"[FATAL]"}, args...)...)
	os.Exit(1)
}
func (l *SimpleLogger) Fatalfln(format string, args ...interface{}) {
	fmt.Printf("[FATAL] "+format+"\n", args...)
	os.Exit(1)
}

func (l *SimpleLogger) Print(args ...interface{}) {
	fmt.Println(args...)
}
func (l *SimpleLogger) Printf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
func (l *SimpleLogger) Println(args ...interface{}) {
	fmt.Println(args...)
}

func (l *SimpleLogger) Sub(name string) maulogger.Logger {
	return l
}

func (l *SimpleLogger) Subm(name string, fields map[string]interface{}) maulogger.Logger {
	return l
}

func (l *SimpleLogger) WithDefaultLevel(level maulogger.Level) maulogger.Logger {
	return l
}

func (l *SimpleLogger) Writer(level maulogger.Level) io.WriteCloser {
	return os.Stdout
}

func (l *SimpleLogger) GetParent() maulogger.Logger {
	return nil
}

func (l *SimpleLogger) SetLevel(level maulogger.Level) {
	// no-op
}

func (l *SimpleLogger) Log(level maulogger.Level, args ...interface{}) {
	fmt.Println(append([]interface{}{fmt.Sprintf("[LOG-%v]", level)}, args...)...)
}

func (l *SimpleLogger) Logf(level maulogger.Level, format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("[LOG-%v] ", level)+format+"\n", args...)
}

func (l *SimpleLogger) Logln(level maulogger.Level, args ...interface{}) {
	fmt.Println(append([]interface{}{fmt.Sprintf("[LOG-%v]", level)}, args...)...)
}

func (l *SimpleLogger) Logfln(level maulogger.Level, format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("[LOG-%v] ", level)+format+"\n", args...)
}

func main() {
	var logger = &SimpleLogger{}

	token := os.Getenv("GROUPME_TOKEN")
	if token == "" {
		if len(os.Args) > 1 {
			token = os.Args[1]
		}
	}

	if token == "" {
		fmt.Print("Enter your GroupMe Developer Token: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		token = strings.TrimSpace(input)
	}

	if token == "" {
		fmt.Println("No token provided. Please set GROUPME_TOKEN environment variable or pass it as an argument.")
		os.Exit(1)
	}

	fmt.Println("Authenticating with GroupMe...")
	// NewClient expects a logger that matches the interface.
	client := groupmeext.NewClient(token, logger)

	fmt.Println("Fetching most recent groups...")
	// Fetch top 5 groups
	// We use IndexGroups instead of IndexChats to be safe with known fields
	groups, err := client.IndexGroups(context.Background(), &groupme.GroupsQuery{
		PerPage: 5,
	})
	if err != nil {
		fmt.Println("Failed to fetch groups:", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d groups. Processing...\n", len(groups))

	for i, group := range groups {
		if i >= 5 {
			break
		}

		fmt.Printf("Processing Group %d/%d: %s (ID: %s)\n", i+1, len(groups), group.Name, group.ID)

		// Fetch messages
		// IndexMessages works for groups
		msgs, err := client.IndexMessages(context.Background(), group.ID, &groupme.IndexMessagesQuery{
			Limit: 10,
		})

		if err != nil {
			fmt.Println("Failed to fetch messages for group:", group.Name, err)
			continue
		}

		fmt.Printf("  Fetched %d messages:\n", len(msgs.Messages))
		for _, msg := range msgs.Messages {
			text := msg.Text
			if text == "" {
				text = "[No Text/Media]"
			}
			// Truncate long text for display
			if len(text) > 50 {
				text = text[:47] + "..."
			}
			fmt.Printf("    - [%s] %s: %s\n", msg.ID, msg.Name, text)
		}
	}

	fmt.Println("------------------------------------------------")
	fmt.Println("Fetching most recent DMs (Chats)...")
	dms, err := client.IndexChats(context.Background(), &groupme.IndexChatsQuery{
		PerPage: 5,
	})
	if err != nil {
		fmt.Println("Failed to fetch DMs:", err)
	} else {
		fmt.Printf("Found %d DMs. Processing...\n", len(dms))
		for i, dm := range dms {
			if i >= 5 {
				break
			}
			fmt.Printf("Processing DM %d/%d: With %s (ID: %s)\n", i+1, len(dms), dm.OtherUser.Name, dm.OtherUser.ID)

			// Fetch messages for DM
			// DMs might use IndexDirectMessages or similar?
			// client.go uses IndexDirectMessages for LoadMessagesAfter/Before in private mode.
			// Let's verify that.
			// Ideally we use what the bridge uses.
			// The bridge uses client.LoadMessagesBefore/After with private=true, which calls IndexDirectMessages.
			// We can try IndexDirectMessages here.

			msgs, err := client.IndexDirectMessages(context.Background(), dm.OtherUser.ID.String(), &groupme.IndexDirectMessagesQuery{
				// Limit not supported?
			})
			if err != nil {
				fmt.Println("Failed to fetch messages for DM:", err)
				continue
			}

			fmt.Printf("  Fetched %d messages:\n", len(msgs.Messages))
			for _, msg := range msgs.Messages {
				text := msg.Text
				if text == "" {
					text = "[No Text/Media]"
				}
				if len(text) > 50 {
					text = text[:47] + "..."
				}
				fmt.Printf("    - [%s] %s: %s\n", msg.ID, msg.Name, text)
			}
		}
	}

	fmt.Println("Done.")
}
