package descriptor_test

import (
	"testing"

	"go.yorun.ai/skel/descriptor"
)

func TestServiceMethodBySkelName(t *testing.T) {
	wireMethod := new(descriptor.Method{
		Name:     "GetUser",
		SkelName: "getUser",
	})
	localMethod := new(descriptor.Method{
		Name:     "getUser",
		SkelName: "other",
	})
	service := new(descriptor.Service{
		Methods: []*descriptor.Method{localMethod, wireMethod},
	})
	if service.MethodBySkelName("getUser") != wireMethod {
		t.Fatal("wire lookup must match SkelName, not the local Name")
	}
	if service.Method("getUser") != localMethod {
		t.Fatal("local lookup must continue matching Name")
	}
	if service.MethodBySkelName("missing") != nil {
		t.Fatal("missing wire method resolved")
	}
}
