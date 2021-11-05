package metering

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestMetering(t *testing.T) {
	logrus.SetOutput(GinkgoWriter)
	logrus.SetLevel(logrus.DebugLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metering Suite")
}
