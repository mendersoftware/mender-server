// Copyright 2026 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

package utils

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CopyOption func(opt *copyOptions)

type copyOptions struct {
	findOpts []options.Lister[options.FindOptions]
}

func WithFindOptions(findOpts ...options.Lister[options.FindOptions]) CopyOption {
	return func(opt *copyOptions) {
		opt.findOpts = append(opt.findOpts, findOpts...)
	}
}

func CopyCollection[S, T any](
	ctx context.Context,
	dst, src *mongo.Collection,
	transform func(doc T) S,
	opts ...CopyOption,
) error {
	var opt copyOptions
	for _, f := range opts {
		f(&opt)
	}
	cur, err := src.Find(ctx, bson.D{}, opt.findOpts...)
	if err != nil {
		return fmt.Errorf("failed to copy documents: %w", err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var doc T
		err := cur.Decode(&doc)
		if err != nil {
			return fmt.Errorf("failed to decode document: %w", err)
		}
		_, err = dst.InsertOne(ctx, transform(doc))
		if err != nil {
			return fmt.Errorf("failed to copy document: %w", err)
		}
	}
	return nil
}

func CopyUpsertCollection[T any](
	ctx context.Context,
	dst, src *mongo.Collection,
	dstQuery func(doc T) any,
	dstUpdate func(doc T) any,
	opts ...CopyOption,
) error {
	var opt copyOptions
	for _, f := range opts {
		f(&opt)
	}
	cur, err := src.Find(ctx, bson.D{}, opt.findOpts...)
	if err != nil {
		return fmt.Errorf("failed to copy documents: %w", err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var doc T
		err := cur.Decode(&doc)
		if err != nil {
			return fmt.Errorf("failed to decode document: %w", err)
		}
		_, err = dst.UpdateOne(ctx,
			dstQuery(doc),
			dstUpdate(doc),
			options.UpdateOne().
				SetUpsert(true))
		if err != nil {
			return fmt.Errorf("failed to copy document: %w", err)
		}
	}
	return nil
}
