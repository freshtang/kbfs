// Copyright 2018 Keybase Inc. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigV2Default(t *testing.T) {
	config := DefaultV2()
	read, list,
		possibleRead, possibleList,
		realm, err := config.GetPermissions("/", nil)
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/", realm)
}

func TestConfigV2Invalid(t *testing.T) {
	err := (&V2{
		Common: Common{
			Version: Version2Str,
		},
		ACLs: map[string]AccessControlV1{
			"/": AccessControlV1{
				WhitelistAdditionalPermissions: map[string]string{
					"alice": PermRead,
				},
			},
		},
	}).EnsureInit()
	require.Error(t, err)
	require.IsType(t, ErrUndefinedUsername{}, err)

	err = (&V2{
		Common: Common{
			Version: Version2Str,
		},
		ACLs: map[string]AccessControlV1{
			"/": AccessControlV1{
				AnonymousPermissions: "",
			},
			"": AccessControlV1{
				AnonymousPermissions: PermRead,
			},
		},
	}).EnsureInit()
	require.Error(t, err)
	require.IsType(t, ErrDuplicateAccessControlPath{}, err)

	err = (&V2{
		Common: Common{
			Version: Version2Str,
		},
		ACLs: map[string]AccessControlV1{
			"/foo": AccessControlV1{
				AnonymousPermissions: "",
			},
			"/foo/../foo": AccessControlV1{
				AnonymousPermissions: PermRead,
			},
		},
	}).EnsureInit()
	require.Error(t, err)
	require.IsType(t, ErrDuplicateAccessControlPath{}, err)

	err = (&V2{
		Common: Common{
			Version: Version2Str,
		},
		ACLs: map[string]AccessControlV1{
			"/": AccessControlV1{
				AnonymousPermissions: "huh?",
			},
		},
	}).EnsureInit()
	require.Error(t, err)
	require.IsType(t, ErrInvalidPermissions{}, err)
}

func TestConfigV2Full(t *testing.T) {
	config := V2{
		Common: Common{
			Version: Version2Str,
		},
		Users: map[string]string{
			"alice": GenerateSHA256PasswordHash("12345"),
			"bob":   GenerateSHA256PasswordHash("54321"),
		},
		ACLs: map[string]AccessControlV1{
			"/": AccessControlV1{
				AnonymousPermissions: "read,list",
			},
			"/alice-and-bob": AccessControlV1{
				WhitelistAdditionalPermissions: map[string]string{
					"alice": PermReadAndList,
					"bob":   PermRead,
				},
			},
			"/bob": AccessControlV1{
				AnonymousPermissions: "",
				WhitelistAdditionalPermissions: map[string]string{
					"bob": PermReadAndList,
				},
			},
			"/public": AccessControlV1{
				AnonymousPermissions: PermReadAndList,
			},
			"/public/not-really": AccessControlV1{
				AnonymousPermissions: "",
				WhitelistAdditionalPermissions: map[string]string{
					"alice": PermReadAndList,
				},
			},
			"/bob/dir/deep-dir/deep-deep-dir": AccessControlV1{},
		},
	}

	ctx := context.Background()

	authenticated := config.Authenticate(ctx, "alice", "12345")
	require.True(t, authenticated)

	read, list, possibleRead, possibleList,
		realm, err := config.GetPermissions("/", nil)
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/", stringPtr("alice"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/", stringPtr("bob"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/", realm)

	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/alice-and-bob", nil)
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/alice-and-bob", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/alice-and-bob", stringPtr("alice"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/alice-and-bob", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/alice-and-bob", stringPtr("bob"))
	require.NoError(t, err)
	require.True(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/alice-and-bob", realm)

	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob", nil)
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob", stringPtr("alice"))
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob", stringPtr("bob"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)

	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/public", nil)
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/public", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/public", stringPtr("alice"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/public", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/public", stringPtr("bob"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/public", realm)

	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/public/not-really", nil)
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/public/not-really", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/public/not-really", stringPtr("alice"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/public/not-really", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/public/not-really", stringPtr("bob"))
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/public/not-really", realm)

	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir", nil)
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir", stringPtr("alice"))
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir", stringPtr("bob"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)

	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir/sub", nil)
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir/sub", stringPtr("alice"))
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir/sub", stringPtr("bob"))
	require.NoError(t, err)
	require.True(t, read)
	require.True(t, list)
	require.True(t, possibleRead)
	require.True(t, possibleList)
	require.Equal(t, "/bob", realm)

	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir/deep-dir/deep-deep-dir", nil)
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.False(t, possibleRead)
	require.False(t, possibleList)
	require.Equal(t, "/bob/dir/deep-dir/deep-deep-dir", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir/deep-dir/deep-deep-dir", stringPtr("alice"))
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.False(t, possibleRead)
	require.False(t, possibleList)
	require.Equal(t, "/bob/dir/deep-dir/deep-deep-dir", realm)
	read, list, possibleRead, possibleList,
		realm, err = config.GetPermissions("/bob/dir/deep-dir/deep-deep-dir", stringPtr("bob"))
	require.NoError(t, err)
	require.False(t, read)
	require.False(t, list)
	require.False(t, possibleRead)
	require.False(t, possibleList)
	require.Equal(t, "/bob/dir/deep-dir/deep-deep-dir", realm)
}