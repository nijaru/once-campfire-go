package presentation

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/rails"
)

func TestVerifiedMentionsPreserveSupportedSGIDs(t *testing.T) {
	raw, err := os.ReadFile("../../reference/vectors/rails_compat.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Secret string `json:"secret_key_base"`
		SGIDs  struct {
			Verify []struct {
				Case, SGID, Purpose, Now string
				Expected                 *string
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	secrets, err := rails.NewSecrets(fixture.Secret)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range fixture.SGIDs.Verify {
		if test.Purpose != "attachable" {
			continue
		}
		t.Run(test.Case, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339Nano, test.Now)
			if err != nil {
				t.Fatal(err)
			}
			var want int64
			if test.Expected != nil {
				want = rails.UserGIDID(*test.Expected)
			}
			id := VerifiedMentionID(secrets, test.SGID, now)
			if id != want {
				t.Fatalf("prepared recipient %d, want %d", id, want)
			}
			users := map[int64]database.UserDisplay{}
			if want != 0 {
				users[want] = database.UserDisplay{ID: want, Name: "User"}
			}
			ctx := MentionContext(secrets, "campfire.test", now, nil, users)
			mention, err := ctx.Resolve(test.SGID, true)
			if err != nil {
				t.Fatal(err)
			}
			if want == 0 {
				if mention != nil {
					t.Fatal("accepted invalid recipient", mention)
				}
			} else if mention == nil || mention.ID != want {
				t.Fatal("lost supported recipient", mention)
			}
		})
	}
}
