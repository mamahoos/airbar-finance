package idempotency

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protoMarshal = protojson.MarshalOptions{
	UseProtoNames:   true,
	EmitUnpopulated: false,
}

var protoUnmarshal = protojson.UnmarshalOptions{
	DiscardUnknown: true,
}

func ResponseToSnapshot(fullMethod string, msg proto.Message) (map[string]any, error) {
	return responseToSnapshot(fullMethod, msg)
}

func SnapshotToResponse(fullMethod string, snapshot map[string]any, msg proto.Message) error {
	return snapshotToResponse(fullMethod, snapshot, msg)
}

func responseToSnapshot(fullMethod string, msg proto.Message) (map[string]any, error) {
	raw, err := protoMarshal.Marshal(msg)
	if err != nil {
		return nil, err
	}

	var snapshot map[string]any
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	snapshot[snapshotMethodKey(fullMethod)] = fullMethod
	return snapshot, nil
}

func snapshotToResponse(fullMethod string, snapshot map[string]any, msg proto.Message) error {
	storedMethod, _ := snapshot[snapshotMethodKey(fullMethod)].(string)
	if storedMethod != "" && storedMethod != fullMethod {
		return fmt.Errorf("snapshot method mismatch: %q != %q", storedMethod, fullMethod)
	}

	raw, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return protoUnmarshal.Unmarshal(raw, msg)
}
