// +build all gy

/*
 * Copyright 2020 The Magma Authors.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package integration

import (
	"fmt"
	"math"
	"testing"
	"time"

	cwfprotos "magma/cwf/cloud/go/protos"
	fegprotos "magma/feg/cloud/go/protos"
	"magma/lte/cloud/go/services/policydb/obsidian/models"

	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
)

func TestGyReAuth(t *testing.T) {
	fmt.Println("\nRunning TestGyReAuth...")

	tr := NewTestRunner(t)
	ruleManager, err := NewRuleManager()
	assert.NoError(t, err)
	defer func() {
		// Clear hss, ocs, and pcrf
		assert.NoError(t, ruleManager.RemoveInstalledRules())
		assert.NoError(t, tr.CleanUp())
	}()

	ues, err := tr.ConfigUEs(1)
	assert.NoError(t, err)

	err = setNewOCSConfig(
		&fegprotos.OCSConfig{
			MaxUsageOctets: &fegprotos.Octets{TotalOctets: ReAuthMaxUsageBytes},
			MaxUsageTime:   ReAuthMaxUsageTimeSec,
			ValidityTime:   ReAuthValidityTime,
		},
	)
	assert.NoError(t, err)

	imsi := ues[0].GetImsi()
	setCreditOnOCS(
		&fegprotos.CreditInfo{
			Imsi:        imsi,
			ChargingKey: 1,
			Volume:      &fegprotos.Octets{TotalOctets: 500 * KiloBytes},
			UnitType:    fegprotos.CreditInfo_Bytes,
		},
	)
	// Set a pass all rule to be installed by pcrf with a monitoring key to trigger updates
	err = ruleManager.AddUsageMonitor(imsi, "mkey-ocs", 500*KiloBytes, 100*KiloBytes)
	assert.NoError(t, err)
	err = ruleManager.AddStaticPassAllToDB("static-pass-all-ocs1", "mkey-ocs", 0, models.PolicyRuleConfigTrackingTypeONLYPCRF, 20)
	assert.NoError(t, err)

	// set a pass all rule to be installed by ocs with a rating group 1
	ratingGroup := uint32(1)
	err = ruleManager.AddStaticPassAllToDB("static-pass-all-ocs2", "", ratingGroup, models.PolicyRuleConfigTrackingTypeONLYOCS, 10)
	assert.NoError(t, err)
	tr.WaitForPoliciesToSync()

	// Apply a dynamic rule that points to the static rules above
	err = ruleManager.AddRulesToPCRF(imsi, []string{"static-pass-all-ocs1", "static-pass-all-ocs2"}, nil)
	assert.NoError(t, err)

	tr.AuthenticateAndAssertSuccess(imsi)

	// Generate over 80% of the quota to trigger a CCR Update
	req := &cwfprotos.GenTrafficRequest{
		Imsi:    imsi,
		Volume:  &wrappers.StringValue{Value: "400K"},
		Bitrate: &wrappers.StringValue{Value: "1M"}}
	_, err = tr.GenULTraffic(req)
	assert.NoError(t, err)
	tr.WaitForEnforcementStatsToSync()

	// Check that UE mac flow is installed and traffic is less than the quota
	tr.AssertPolicyUsage(imsi, "static-pass-all-ocs2", 0, 5*MegaBytes+Buffer)

	// Top UP extra credits (2.5M total)
	err = setCreditOnOCS(
		&fegprotos.CreditInfo{
			Imsi:        imsi,
			ChargingKey: ratingGroup,
			Volume:      &fegprotos.Octets{TotalOctets: 2 * MegaBytes},
			UnitType:    fegprotos.CreditInfo_Bytes,
		},
	)
	assert.NoError(t, err)

	// Send ReAuth Request to update quota
	raa, err := sendChargingReAuthRequest(imsi, ratingGroup)
	assert.NoError(t, err)
	assert.Eventually(t, tr.WaitForChargingReAuthToProcess(raa, imsi), time.Minute, 2*time.Second)

	// Check ReAuth success
	assert.Equal(t, diam.LimitedSuccess, int(raa.ResultCode))

	// Generate over 1M of data to check that initial quota was updated
	req = &cwfprotos.GenTrafficRequest{Imsi: imsi,
		Volume: &wrappers.StringValue{Value: "1M"},
	}
	_, err = tr.GenULTraffic(req)
	assert.NoError(t, err)
	tr.WaitForEnforcementStatsToSync()

	// Check that initial quota was exceeded
	tr.AssertPolicyUsage(imsi, "static-pass-all-ocs2", 400*KiloBytes, uint64(math.Round(2.4*MegaBytes+Buffer)))

	// trigger disconnection
	tr.DisconnectAndAssertSuccess(imsi)
	fmt.Println("wait for flows to get deactivated")
	time.Sleep(3 * time.Second)
}
