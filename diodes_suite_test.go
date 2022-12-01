package diodes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDiodes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diodes Suite")
}
