package log_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kqlite/kqlite/pkg/util/log"
)

var _ = Describe("logger", func() {
	It("Simple create", func() {
		logger := log.CreateLogger("", log.LogLevelInfo, "")
		logger.Info("Logger created")
		Expect(logger.Enabled()).To(BeTrue())
	})
	It("Set log level", func() {
		logger := log.CreateLogger("", log.LogLevelDebug, "")
		logger.V(1).Info("Logger created with log level 1")
		logger.V(2).Info("Messgae with log level 2")
		Expect(logger.Enabled()).To(BeTrue())
	})
	It("Set log file output", func() {
		logFilename := "logout.log"

		logger := log.CreateLogger("", log.LogLevelInfo, logFilename)
		logger.Info("Logger created")
		Expect(logger.Enabled()).To(BeTrue())
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
