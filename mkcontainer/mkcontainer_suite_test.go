package mkcontainer_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMkcontainer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mkcontainer Suite")
}
