// Copyright 2023 Northern.tech AS
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

package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mendersoftware/mender-server/services/deployments/store"
)

func IteratorFromCursor[T interface{}](cur *mongo.Cursor) store.Iterator[T] {
	return (*iterator[T])(cur)
}

type iterator[T interface{}] mongo.Cursor

func (it *iterator[T]) Next(ctx context.Context) (bool, error) {
	cur := (*mongo.Cursor)(it)
	next := cur.Next(ctx)
	return next, cur.Err()
}

func (it *iterator[T]) Decode(value *T) error {
	return (*mongo.Cursor)(it).Decode(value)
}

func (it *iterator[T]) Close(ctx context.Context) error {
	return (*mongo.Cursor)(it).Close(ctx)
}
