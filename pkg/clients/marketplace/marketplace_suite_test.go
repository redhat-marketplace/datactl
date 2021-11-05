package marketplace

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestMarketplace(t *testing.T) {
	logrus.SetOutput(GinkgoWriter)
	logrus.SetLevel(logrus.DebugLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Marketplace Suite")
}
