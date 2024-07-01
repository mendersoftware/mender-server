// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package mongo

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
	"github.com/mendersoftware/mender-server/pkg/identity"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

func TestPopControlMessage(t *testing.T) {
	ErrMessageContentsDoNotMatch := errors.New("message contents do not match")
	testCases := []struct {
		Name string

		Ctx         context.Context
		SessionID   string
		ControlData string
		Messages    []*app.Control

		Error error
	}{
		{
			Name: "ok delay message",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000000",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000000",
			//            0    1   2   3   4   5   6
			//echo -n -e '\x02\x20\x00\x00\x00\x08\x00' | gzip - | base64
			ControlData: "H4sIALIDNWAAA2NSYGBg4GAAAGlBhsUHAAAA",
			Messages: []*app.Control{
				{
					Type:           app.DelayMessage, // offset0 \x02
					Offset:         32,               // offset 1-4 \x20\x00
					DelayMs:        8,                // offset 5-6 \x08\x00
					TerminalWidth:  0,
					TerminalHeight: 0,
				},
			},
		},
		{
			Name: "ok resize message",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000000",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000000",
			//            0    1   2   3   4   5   6
			//echo -n -e '\x01\x20\x00\x00\x00\x50\x00\x18\x00' | gzip - | base64
			ControlData: "H4sIAHcFNWAAA2NUYGBgCGCQYAAAQPUSQQkAAAA=",
			Messages: []*app.Control{
				{
					Type:           app.ResizeMessage, // offset0 \x02
					Offset:         32,                // offset 1-4 \x20\x00
					DelayMs:        0,                 //not present in the resize message
					TerminalWidth:  80,                // offset 5-6 \x08\x00
					TerminalHeight: 24,                // offset 7-8 \x08\x00
				},
			},
		},
		{
			Name: "ok more than one",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000010",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000010",
			//            0    1   2   3   4
			//echo -n -e '\x02\x20\x00\x00\x00\x08\x00\x02\xfa\x3f\x00\x00\x00\x80\x01\x01\x04\x00\x00\x50\x00\x18\x00' | gzip - | base64
			ControlData: "H4sIAENINWAAA2NSYGBg4GBg+mUPpBsYGVkYGAIYJBgAdudD1RcAAAA=",
			Messages: []*app.Control{
				{
					Type:           app.DelayMessage, // offset 0 \x02
					Offset:         32,               // offset 1-4 \x20\x00
					DelayMs:        8,                // offset 5-6 \x08\x00
					TerminalWidth:  0,
					TerminalHeight: 0,
				},
				{
					Type:           app.DelayMessage, // offset 0 \x02
					Offset:         16378,            // offset 1-4 \x20\x00
					DelayMs:        32768,            // offset 5-6 \x08\x00
					TerminalWidth:  0,
					TerminalHeight: 0,
				},
				{
					Type:           app.ResizeMessage, // offset 0 \x01
					Offset:         1025,              // offset 1-5 \x20\x00
					DelayMs:        0,                 // does not exist in ResizeMessage
					TerminalWidth:  80,                // offset 5-6
					TerminalHeight: 24,                // offset 7-8
				},
			},
		},
		{
			Name: "error delay message not matching",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000000",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000000",
			//            0    1   2   3   4   5   6
			//echo -n -e '\x02\x20\x00\x00\x00\x08\x00' | gzip - | base64
			ControlData: "H4sIALIDNWAAA2NSYGBg4GAAAGlBhsUHAAAA",
			Messages: []*app.Control{
				{
					Type:           app.DelayMessage, // offset0 \x02
					Offset:         32,               // offset 1-4 \x20\x00
					DelayMs:        9,                // offset 5-6 \x08\x00
					TerminalWidth:  0,
					TerminalHeight: 0,
				},
			},
			Error: ErrMessageContentsDoNotMatch,
		},
		{
			Name: "error resize message not matching",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000000",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000000",
			//            0    1   2   3   4   5   6
			//echo -n -e '\x01\x20\x00\x00\x00\x50\x00\x18\x00' | gzip - | base64
			ControlData: "H4sIAHcFNWAAA2NUYGBgCGCQYAAAQPUSQQkAAAA=",
			Messages: []*app.Control{
				{
					Type:           app.ResizeMessage, // offset0 \x02
					Offset:         32,                // offset 1-4 \x20\x00
					DelayMs:        0,                 //not present in the resize message
					TerminalWidth:  81,                // offset 5-6 \x08\x00
					TerminalHeight: 24,                // offset 7-8 \x08\x00
				},
			},
			Error: ErrMessageContentsDoNotMatch,
		},
		{
			Name: "error unknown message type",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000001",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000000",
			//            0    1   2   3   4
			//echo -n -e '\x22\x20\x00\x08\x00' | gzip - | base64
			ControlData: "H4sIAHurLWAAA1NSYOBgAABPrsgVBQAAAA==",
		},
		{
			Name: "error buffer does not contain a full message",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000001",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000002",
			//            0    1   2   3   4
			//echo -n -e '\x02\x20\x00\x08' | gzip - | base64
			ControlData: "H4sIAOisLWAAA2NSYOAAAEXZ270EAAAA",
		},
		{
			Name: "error buffer too short to contain any message",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000001",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000002",
			//            0    1   2   3   4
			//echo -n -e '\x02\x20' | gzip - | base64
			ControlData: "H4sIAHuuLWAAA2NSAAC1UIFIAgAAAA==",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			db.Wipe()
			ds := &DataStoreMongo{client: db.Client()}
			defer ds.DropDatabase()

			database := db.Client().Database(mstore.DbNameForTenant(
				"000000000000000000000000", DbName,
			))
			collSess := database.Collection(ControlCollectionName)
			d, err := base64.StdEncoding.DecodeString(tc.ControlData)
			assert.NoError(t, err)

			_, err = collSess.InsertOne(nil, &model.ControlData{
				ID:        uuid.New(),
				SessionID: tc.SessionID,
				Control:   d,
				CreatedTs: time.Now().UTC(),
				ExpireTs:  time.Now().UTC(),
			})
			assert.NoError(t, err)

			c, err := collSess.Find(context.Background(), bson.M{
				dbFieldSessionID: tc.SessionID,
			})
			assert.NotNil(t, c)
			assert.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
			defer cancel()

			r := NewControlMessageReader(ctx, c)
			assert.NotNil(t, r)

			if len(tc.Messages) == 0 {
				m := r.Pop()
				assert.Nil(t, m)
			} else {
				actualMessages := make([]*app.Control, len(tc.Messages))
				for i, _ := range tc.Messages {
					actualMessages[i] = r.Pop()
				}
				if tc.Error == ErrMessageContentsDoNotMatch {
					assert.NotEqual(t, tc.Messages, actualMessages)
				} else {
					assert.Equal(t, tc.Messages, actualMessages)
					m := r.Pop()
					assert.Nil(t, m)
				}
			}
		})
	}
}
