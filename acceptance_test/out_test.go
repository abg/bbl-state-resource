package acceptance_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("out", func() {
	var (
		upTargetDir string
		upOutInput  *bytes.Buffer

		downTargetDir string
		downOutInput  *bytes.Buffer
	)

	BeforeEach(func() {
		upRequest := fmt.Sprintf(`{
			"source": {
				"name": "out-test-test-env",
				"iaas": "gcp",
				"gcp-region": "us-east1",
				"gcp-service-account-key": %s
			},
			"params": {
				"command": "up"
			}
		}`, strconv.Quote(serviceAccountKey))

		var err error
		upTargetDir, err = ioutil.TempDir("", "up_out_test")
		Expect(err).NotTo(HaveOccurred())
		upOutInput = bytes.NewBuffer([]byte(upRequest))

		downRequest := fmt.Sprintf(`{
			"source": {
				"name": "out-test-test-env",
				"iaas": "gcp",
				"gcp-region": "us-east1",
				"gcp-service-account-key": %s
			},
			"params": {
				"command": "down"
			}
		}`, strconv.Quote(serviceAccountKey))

		downTargetDir, err = ioutil.TempDir("", "down_out_test")
		Expect(err).NotTo(HaveOccurred())
		downOutInput = bytes.NewBuffer([]byte(downRequest))
	})

	AfterEach(func() {
		defer os.RemoveAll(upTargetDir)   // ignore the error
		defer os.RemoveAll(downTargetDir) // ignore the error
		By("bbling down", func() {
			cmd := exec.Command(outBinaryPath, downTargetDir)
			cmd.Stdin = downOutInput
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say(`{"version":{"ref":"[0-9a-f]+"}}`))
			_, err = os.Stat(filepath.Join(downTargetDir, "bbl-state.json"))
			Expect(err).To(HaveOccurred())
		})
	})

	It("bbls up and down successfully from different dirs configured with the same source", func() {
		By("bbling up", func() {
			cmd := exec.Command(outBinaryPath, upTargetDir)
			cmd.Stdin = upOutInput
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say(`{"version":{"ref":"[0-9a-f]+"}}`))
			_, err = os.Open(filepath.Join(upTargetDir, "bbl-state.json"))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
