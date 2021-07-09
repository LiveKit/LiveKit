// +build tools

package tools

import (
	_ "github.com/google/wire/cmd/wire"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "github.com/twitchtv/twirp/protoc-gen-twirp"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
