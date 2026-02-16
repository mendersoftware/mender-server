// Copyright 2025 Northern.tech AS
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
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	mongostore "github.com/mendersoftware/mender-server/pkg/mongo"

	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

const (
	CollNameLog = "log"

	KeyEventTs       = "event_ts"
	KeyEventExpireTs = "expire_ts"
)

var (
	eventExpiration int64

	ErrFailedToGetEvents = errors.New("failed to get events")
)

func SetEventExpiration(exp int64) {
	eventExpiration = exp
}

func (db *DataStoreMongo) GetEvents(
	ctx context.Context,
	fltr model.EventsFilter,
) ([]model.Event, error) {
	var (
		err     error
		results = []model.Event{}
	)

	collEvents := db.Collection(CollNameLog)
	findOpts := mopts.Find().
		SetSort(bson.D{{Key: KeyEventTs, Value: -1}}).
		SetSkip(fltr.Skip)
	if fltr.Limit > 0 {
		findOpts.SetLimit(fltr.Limit)
	}

	filter := bson.D{}
	if fltr.IntegrationID != nil {
		filter = append(filter,
			bson.E{
				Key:   KeyStatus + "." + KeyIntegrationID,
				Value: uuid.MustParse(*fltr.IntegrationID),
			},
		)
	}
	cur, err := collEvents.Find(ctx,
		mongostore.WithTenantID(ctx, filter),
		findOpts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error executing log collection request")
	}
	if err = cur.All(ctx, &results); err != nil {
		return nil, errors.Wrap(err, "error retrieving log collection results")
	}

	return results, nil
}

func (db *DataStoreMongo) SaveEvent(
	ctx context.Context,
	event model.Event,
) error {
	now := time.Now()
	event.EventTS = now
	event.ExpireTS = now.Add(time.Second * time.Duration(eventExpiration))

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	collEvents := db.Collection(CollNameLog)

	_, err := collEvents.
		InsertOne(ctx, mongostore.WithTenantID(ctx, event))
	if err != nil {
		return errors.Wrapf(err, "failed to store an event %v", event)
	}

	return nil
}
