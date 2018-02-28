package acceptance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = FDescribe("check", func() {
	Context("when there is something in gcp", func() {
		var (
			bblStateContents string
			check            *bytes.Buffer
			version          concourse.Version
		)
		BeforeEach(func() {
			checkRequest := fmt.Sprintf(`{
				"source": {
					"name": "%s-check-test-test-env",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"version": {"ref": "the-greatest"}
			}`, projectId, strconv.Quote(serviceAccountKey))

			var req concourse.InRequest
			err := json.Unmarshal([]byte(checkRequest), &req)
			Expect(err).NotTo(HaveOccurred())
			// this client isn't well tested, so we're going
			// to violate some abstraction layers to test it here
			// against the real api
			client, err := storage.NewStorageClient(req.Source)
			Expect(err).NotTo(HaveOccurred())

			By("uploading a bogus bbl state with some unique contents", func() {
				uploadDir, err := ioutil.TempDir("", "upload_dir")
				Expect(err).NotTo(HaveOccurred())
				filename := filepath.Join(uploadDir, "bbl-state.json")
				f, err := os.Create(filename)
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()

				bblStateContents = fmt.Sprintf(`{"randomDir": "%s"}`, uploadDir)
				_, err = f.Write([]byte(bblStateContents))
				Expect(err).NotTo(HaveOccurred())

				version, err = client.Upload(uploadDir)
				Expect(err).NotTo(HaveOccurred())
			})

			check = bytes.NewBuffer([]byte(checkRequest))
		})

		It("prints the latest version of the resource", func() {
			cmd := exec.Command(checkBinaryPath)
			cmd.Stdin = check
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf(`\[{"ref":"%s"}\]`, version.Ref)))
		})
	})

	Context("when there is nothing stored in gcp", func() {
		It("prints an empty json list", func() {
			checkRequest := fmt.Sprintf(`{
				"source": {
					"name": "%s-empty-bucket-check-test",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"version": {"ref": "the-greatest"}
			}`, projectId, strconv.Quote(serviceAccountKey))

			cmd := exec.Command(checkBinaryPath)
			cmd.Stdin = bytes.NewBuffer([]byte(checkRequest))

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(Equal(`[]`))
		})
	})
})
