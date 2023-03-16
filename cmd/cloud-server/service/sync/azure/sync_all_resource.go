/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package azure

import (
	"time"

	"hcm/pkg/client"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncAllResourceOption ...
type SyncAllResourceOption struct {
	AccountID string `json:"account_id" validate:"required"`
	// SyncPublicResource 是否同步公共资源
	SyncPublicResource bool `json:"sync_public_resource" validate:"omitempty"`
}

// Validate SyncAllResourceOption
func (opt *SyncAllResourceOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SyncAllResource sync resource.
func SyncAllResource(kt *kit.Kit, cliSet *client.ClientSet, opt *SyncAllResourceOption) error {

	if err := opt.Validate(); err != nil {
		return err
	}

	start := time.Now()
	logs.V(3).Infof("azure account[%s] sync all resource start, time: %v, opt: %v, rid: %s", opt.AccountID,
		start, opt, kt.Rid)

	var hitErr error
	defer func() {
		if hitErr != nil {
			// TODO: 更新账号同步状态为同步异常

			return
		}

		logs.V(3).Infof("azure account[%s] sync all resource end, cost: %v, opt: %v, rid: %s", opt.AccountID,
			time.Since(start), opt, kt.Rid)
	}()

	// TODO: 修改账号表中同步状态字段和同步时间字段

	if err := SyncResourceGroup(kt, cliSet.HCService(), opt.AccountID); err != nil {
		return err
	}

	resourceGroupNames, err := ListResourceGroup(kt, cliSet.DataService(), opt.AccountID)
	if err != nil {
		return err
	}

	if opt.SyncPublicResource {
		syncOpt := &SyncPublicResourceOption{
			AccountID:          opt.AccountID,
			ResourceGroupNames: resourceGroupNames,
		}
		if hitErr = SyncPublicResource(kt, cliSet, syncOpt); hitErr != nil {
			logs.Errorf("sync public resource failed, err: %v, opt: %v, rid: %s", hitErr, opt, kt.Rid)
			return hitErr
		}
	}

	if hitErr = SyncDisk(kt, cliSet.HCService(), opt.AccountID, resourceGroupNames); hitErr != nil {
		return hitErr
	}

	if hitErr = SyncVpc(kt, cliSet.HCService(), opt.AccountID, resourceGroupNames); hitErr != nil {
		return hitErr
	}

	if hitErr = SyncSubnet(kt, cliSet.HCService(), cliSet.DataService(), opt.AccountID,
		resourceGroupNames); hitErr != nil {
		return hitErr
	}

	if hitErr = SyncEip(kt, cliSet.HCService(), opt.AccountID, resourceGroupNames); hitErr != nil {
		return hitErr
	}

	if hitErr = SyncSG(kt, cliSet.HCService(), opt.AccountID, resourceGroupNames); hitErr != nil {
		return hitErr
	}

	if hitErr = SyncCvm(kt, cliSet.HCService(), opt.AccountID, resourceGroupNames); hitErr != nil {
		return hitErr
	}

	if hitErr = SyncRouteTable(kt, cliSet.HCService(), opt.AccountID, resourceGroupNames); hitErr != nil {
		return hitErr
	}

	if hitErr = SyncNetworkInterface(kt, cliSet.HCService(), opt.AccountID, resourceGroupNames); hitErr != nil {
		return hitErr
	}

	// TODO: 更新同步状态字段为同步结束，更新结束时间

	return nil
}