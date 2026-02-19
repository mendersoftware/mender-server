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
	"reflect"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type migration_1_2_4 struct {
	client *mongo.Client
	db     string
}

func (m *migration_1_2_4) Up(from migrate.Version) error {
	ctx := context.Background()
	l := log.FromContext(ctx)

	coll := m.client.Database(m.db).Collection(CollectionDeployments)

	// update all deployments with finished timestamp to status "finished"
	_, _ = coll.UpdateMany(ctx, bson.M{
		StorageKeyDeploymentFinished: bson.M{
			"$ne": nil,
		},
	}, bson.M{
		"$set": bson.M{
			StorageKeyDeploymentStatus: model.DeploymentStatusFinished,
		},
	})

	// recalculate stats for all the non-finished deployments
	// we'll be iterating and modifying - sort by _id to ensure every doc is handled exactly once
	fopts := options.Find()
	fopts.SetSort(bson.M{"_id": 1})
	fopts.SetNoCursorTimeout(true)

	cur, err := coll.Find(ctx, bson.M{
		StorageKeyDeploymentFinished: bson.M{
			"$eq": nil,
		},
	}, fopts)
	if err != nil {
		return errors.Wrap(err, "failed to get deployments")
	}
	defer cur.Close(ctx)

	allstats, err := m.aggregateDeviceStatuses(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get aggregated device statuses")
	}

	for cur.Next(ctx) {
		var dep model.Deployment
		err := cur.Decode(&dep)
		if err != nil {
			return errors.Wrap(err, "failed to get deployment")
		}
		l.Infof("processing deployment %s with stats %v", dep.Id, dep.Stats)

		var stats model.Stats
		if allstats[dep.Id] != nil {
			stats = *allstats[dep.Id]
		} else {
			stats = model.NewDeviceDeploymentStats()
		}
		newstats := stats
		l.Infof("computed stats: %v", dep.Stats)

		// substitute stats to recalc status with deployment.GetStatus
		dep.Stats = newstats
		status := m.getStatus(&dep)
		deviceCount, err := m.deviceCountByDeployment(ctx, dep.Id)
		if err != nil {
			return errors.Wrapf(err, "failed to count device deployments for deployment %s", dep.Id)
		}

		sets := bson.M{
			StorageKeyDeploymentStatus:     status,
			StorageKeyDeploymentMaxDevices: deviceCount,
		}

		if !reflect.DeepEqual(newstats, dep.Stats) {
			sets[StorageKeyDeploymentStats] = newstats
		}

		res, err := coll.UpdateOne(ctx, bson.M{"_id": dep.Id}, bson.M{"$set": sets})
		if err != nil {
			return errors.Wrapf(err, "failed to update deployment %s", dep.Id)
		}

		if res.MatchedCount == 0 {
			return errors.Wrapf(err, "can't find deployment for update: %s", dep.Id)
		}

		l.Infof("processing deployment %s: success", dep.Id)
	}

	if err := cur.Err(); err != nil {
		l.Warnf("cursor error after processing: %v", err)
		return err
	}

	// have an index on just the plain Deployment.Status field
	// for easy querying by status
	storage := NewDataStoreMongoWithClient(m.client)
	if err := storage.EnsureIndexes(m.db, CollectionDeployments,
		DeploymentStatusIndex); err != nil {
		return err
	}

	return nil
}

func isFinished(d *model.Deployment) bool {
	if d.Stats[model.DeviceDeploymentStatusPendingStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusDownloadingStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusInstallingStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusRebootingStr] == 0 {
		return true
	}

	return false
}

func isPending(d *model.Deployment) bool {
	//pending > 0, evt else == 0
	if d.Stats[model.DeviceDeploymentStatusPendingStr] > 0 &&
		d.Stats[model.DeviceDeploymentStatusDownloadingStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusInstallingStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusRebootingStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusSuccessStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusAlreadyInstStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusFailureStr] == 0 &&
		d.Stats[model.DeviceDeploymentStatusNoArtifactStr] == 0 {

		return true
	}

	return false
}

func (m *migration_1_2_4) getStatus(deployment *model.Deployment) model.DeploymentStatus {
	if isPending(deployment) {
		return model.DeploymentStatusPending
	} else if isFinished(deployment) {
		return model.DeploymentStatusFinished
	} else {
		return model.DeploymentStatusInProgress
	}
}

// aggregateDeviceStatuses calculates:
// - stats
// - statistics
// for deployment 'depId', based on individual device statuses
// it mirrors store.AggregateDeviceDeploymentByStatus, and freezes
// it's implementation in case it changes/is removed
// note that device statuses are the best bet as a single source of
// truth on deployment status (used for all GETs at the time of writing this migration)
func (m *migration_1_2_4) aggregateDeviceStatuses(
	ctx context.Context,
) (map[string]*model.Stats, error) {
	deviceDeployments := m.client.Database(m.db).Collection(CollectionDevices)

	group := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id",
				Value: bson.D{
					{Key: "deploymentid", Value: "$" + StorageKeyDeviceDeploymentDeploymentID},
					{Key: "status", Value: "$" + StorageKeyDeviceDeploymentStatus},
				},
			},
			{Key: "count",
				Value: bson.M{"$sum": 1}}},
		},
	}
	pipeline := []bson.D{
		group,
	}

	var results []struct {
		ID struct {
			DeploymentID string `bson:"deploymentid"`
			Status       model.
					DeviceDeploymentStatus `bson:"status"`
		} `bson:"_id"`
		Count int
	}

	cursor, err := deviceDeployments.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &results); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	stats := make(map[string]*model.Stats)
	for _, res := range results {
		if stats[res.ID.DeploymentID] == nil {
			raw := model.NewDeviceDeploymentStats()
			stats[res.ID.DeploymentID] = &raw
		}
		(*stats[res.ID.DeploymentID]).Set(res.ID.Status, res.Count)
	}
	return stats, nil
}

func (m *migration_1_2_4) deviceCountByDeployment(ctx context.Context, id string) (int, error) {
	collDevs := m.client.Database(m.db).Collection(CollectionDevices)

	filter := bson.M{
		"deploymentid": id,
	}

	deviceCount, err := collDevs.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(deviceCount), nil
}

func (m *migration_1_2_4) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 4)
}
