// Copyright 2025 Northern.tech AS
//
//	All Rights Reserved

package types

func Pointer[T any](value T) *T { return &value }
