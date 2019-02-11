// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dashboard provides testing of the grafana dashboards used in Istio
// to provide mesh monitoring capabilities.

package maistra

import (
	"io/ioutil"
	"testing"
	"time"

	"istio.io/istio/pkg/log"
	"istio.io/istio/tests/util"
)

func cleanup11(namespace, kubeconfig string) {
	log.Infof("# Cleanup. Following error can be ignored...")
	util.KubeDelete(namespace, bookinfoRatingMySQLYaml, kubeconfig)
	util.KubeDelete(namespace, bookinfoRatingMySQLv2Yaml, kubeconfig)
	util.KubeDelete(namespace, bookinfoRatingMySQLServiceEntryYaml, kubeconfig)
	log.Info("Waiting for rules to be cleaned up. Sleep 10 seconds...")
	time.Sleep(time.Duration(10) * time.Second)
}

func configTCPRatings(namespace, kubeconfig string) error {
	if err := util.KubeApply(namespace, bookinfoRatingMySQLv2Yaml, kubeconfig); err != nil {
		return err
	}
	log.Info("Waiting for rules to be cleaned up. Sleep 5 seconds...")
	time.Sleep(time.Duration(5) * time.Second)
	if err := util.KubeApply(namespace, bookinfoRatingMySQLYaml, kubeconfig); err != nil {
		return err
	}
	log.Info("Waiting for rules to be cleaned up. Sleep 10 seconds...")
	time.Sleep(time.Duration(10) * time.Second)
	return nil
}

func configEgressTCP(namespace, kubeconfig string) error {
	if err := util.KubeApply(namespace, bookinfoRatingMySQLServiceEntryYaml, kubeconfig); err != nil {
		return err
	}
	log.Info("Waiting for rules to be cleaned up. Sleep 30 seconds...")
	time.Sleep(time.Duration(30) * time.Second)
	return nil
}


func Test11(t *testing.T) {
	log.Info("# TC_11 Control Egress TCP Traffic")
	Inspect(configTCPRatings(testNamespace, ""), "failed to apply rules", "", t)
	resp, _, err := GetHTTPResponse(productpageURL, nil)
	Inspect(err, "failed to get productpage", "", t)
	body, err := ioutil.ReadAll(resp.Body)
	Inspect(err, "failed to read response body", "", t)
	Inspect(
		CompareHTTPResponse(body, "productpage-normal-user-rating-unavailable.html"),
		"Didn't get expected response",
		"Success. Response matches expected Ratings unavailable",
		t)
	CloseResponseBody(resp)

	log.Info("# Define a TCP mesh-external service entry")
	Inspect(configEgressTCP(testNamespace, ""), "failed to apply service entry", "", t)

	resp, _, err = GetHTTPResponse(productpageURL, nil)
	Inspect(err, "failed to get productpage", "", t)
	body, err = ioutil.ReadAll(resp.Body)
	Inspect(err, "failed to read response body", "", t)
	Inspect(
		CompareHTTPResponse(body, "productpage-normal-user-rating-one-star.html"),
		"Didn't get expected response",
		"Success. Response matches expected one star Ratings",
		t)
	CloseResponseBody(resp)

	defer cleanup11(testNamespace, "")
	defer func() {
		// recover from panic if one occured. This allows cleanup to be executed after panic.
		if err := recover(); err != nil {
			log.Infof("Test failed: %v", err)
		}
	}()
}