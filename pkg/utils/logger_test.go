package utils_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kqlite/kqlite/pkg/utils"
)

var _ = Describe("logger", func() {
	It("Simple create", func() {
		log := utils.CreateLogger("", utils.LogLevelInfo, "")
		log.Info("Logger created")
		Expect(log.Enabled()).To(BeTrue())
	})
	It("Set log level", func() {
		log := utils.CreateLogger("", utils.LogLevelDebug, "")
		log.V(1).Info("Logger created with log level 1")
		log.V(2).Info("Messgae with log level 2")
		Expect(log.Enabled()).To(BeTrue())
	})
	It("Set log file output", func() {
		logFilename := "logout.log"

		log := utils.CreateLogger("", utils.LogLevelInfo, logFilename)
		log.Info("Logger created")
		Expect(log.Enabled()).To(BeTrue())
		logf, err := os.Open(logFilename)
		Expect(err).NotTo(HaveOccurred())
		b := make([]byte, 256)
		n, err := logf.Read(b)
		Expect(err).NotTo(HaveOccurred())
		Expect(n).NotTo(BeZero())
		Expect(b).NotTo(BeEmpty())

		By("Clean up", func() {
			logf.Close()
			Expect(os.Remove(logFilename)).NotTo(HaveOccurred())
		})
	})
})
