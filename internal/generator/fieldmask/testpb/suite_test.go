package testpb_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGenerated(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fieldmask Testpb Suite")
}
