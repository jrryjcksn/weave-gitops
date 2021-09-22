// +build !unittest

package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops/test/acceptance/test/cluster"
)

func TestAcceptance(t *testing.T) {

	if testing.Short() {
		t.Skip("Skip User Acceptance Tests")
	}

	RegisterFailHandler(Fail)
	//gomega.RegisterFailHandler(GomegaFail)
	RunSpecs(t, "Weave GitOps User Acceptance Tests")
}

var clusterPool *cluster.ClusterPool

var globalCtx context.Context
var globalCancel func()

var _ = SynchronizedBeforeSuite(func() []byte {
	dbDirectory := ""

	if os.Getenv(CI) == "" {
		var err error
		dbDirectory, err = ioutil.TempDir("", "db-directory")
		Expect(err).NotTo(HaveOccurred())

		fmt.Println("context-directory", dbDirectory)

		err = cluster.CreateClusterDB(dbDirectory)
		Expect(err).NotTo(HaveOccurred())

		clusterPool = cluster.NewClusterPool()

		clusterPool.GenerateClusters(dbDirectory, config.GinkgoConfig.ParallelTotal)
		go clusterPool.GenerateClusters(dbDirectory, 1)

		globalCtx, globalCancel = context.WithCancel(context.Background())

		go clusterPool.CreateClusterOnRequest(globalCtx, dbDirectory)
	}

	return []byte(dbDirectory)
}, func(dbDirectory []byte) {

	contextDirectory = dbDirectory

	SetDefaultEventuallyTimeout(EVENTUALLY_DEFAULT_TIMEOUT)
	DEFAULT_SSH_KEY_PATH = os.Getenv("HOME") + "/.ssh/id_rsa"
	GITHUB_ORG = os.Getenv("GITHUB_ORG")
	WEGO_BIN_PATH = os.Getenv("WEGO_BIN_PATH")
	if WEGO_BIN_PATH == "" {
		WEGO_BIN_PATH = "/usr/local/bin/gitops"
	}
	log.Infof("GITOPS Binary Path: %s", WEGO_BIN_PATH)
})

func GomegaFail(message string, callerSkip ...int) {
	if webDriver != nil {
		filepath := takeScreenshot()
		fmt.Printf("Failure screenshot is saved in file %s\n", filepath)
	}

	ginkgo.Fail(message, callerSkip...)
}

var _ = SynchronizedAfterSuite(func() {
}, func() {
	if os.Getenv(CI) == "" {
		globalCancel()
		clusterPool.End()
		cmd := "kind delete clusters --all"
		c := exec.Command("sh", "-c", cmd)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err := c.Run()
		if err != nil {
			fmt.Printf("Error deleting ramaining clusters %s\n", err)
		}
		err = os.RemoveAll(string(contextDirectory))
		if err != nil {
			fmt.Printf("Error deleting root folder %s\n", err)
		}
	}

})
