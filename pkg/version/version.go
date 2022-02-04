package version

import (
	"fmt"
	"runtime"
	"time"

	commonv1 "go.buf.build/odpf/gw/odpf/proton/odpf/common/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	Version   = "dev"
	Commit    = "dev"
	BuildTime = ""
)

func GetVersionAndBuildInfo() *commonv1.Version {
	v := &commonv1.Version{
		Version:      Version,
		Commit:       Commit,
		LangVersion:  runtime.Version(),
		Os:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	if bt, err := time.Parse(time.RFC3339, BuildTime); err == nil {
		v.BuildTime = timestamppb.New(bt)
	}

	return v
}

func Print() error {
	_, err := fmt.Println(protojson.Format(GetVersionAndBuildInfo()))
	return err
}
