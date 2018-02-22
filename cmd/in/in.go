package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr,
			"not enough args - usage: %s <target directory>\n",
			os.Args[0],
		)
		os.Exit(1)
	}

	rawBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read configuration: %s\n", err)
		os.Exit(1)
	}

	inRequest, err := concourse.NewInRequest(rawBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %s\n", err)
		os.Exit(1)
	}

	storageClient, err := storage.NewStorageClient(inRequest.Source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create storage client: %s", err.Error())
		os.Exit(1)
	}

	version, err := storageClient.Download(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to download bbl state: %s", err.Error())
		os.Exit(1)
	}

	err = json.NewEncoder(os.Stdout).Encode(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal version: %s", err.Error())
		os.Exit(1)
	}
}
