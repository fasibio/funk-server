package main

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
)

func Test_genID(t *testing.T) {
	res := genID()
	if len(res) != 27 {
		t.Errorf("genUID have to generate a %v sing long uid but got %v", 18, len(res))
	}
}

func Test_getIlmPolicyBody(t *testing.T) {
	cupaloy.SnapshotT(t, getIlmPolicyBody("20d"))
}

func Test_getPolicyTemplateBody(t *testing.T) {
	cupaloy.SnapshotT(t, getPolicyTemplateBody())
}
