package diodes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDiodes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diodes Suite")
}
