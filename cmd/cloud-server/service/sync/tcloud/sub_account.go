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

package tcloud

import (
	"time"

	"hcm/pkg/api/hc-service/sync"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncSubAccount sync sub account
func SyncSubAccount(kt *kit.Kit, hcCli *hcservice.Client, accountID string) error {

	start := time.Now()
	logs.V(3).Infof("tcloud account[%s] sync sub account start, time: %v, rid: %s", accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("tcloud account[%s] sync sub account end, cost: %v, rid: %s", accountID,
			time.Since(start), kt.Rid)
	}()

	req := &sync.TCloudGlobalSyncReq{
		AccountID: accountID,
	}
	if err := hcCli.TCloud.Account.SyncSubAccount(kt, req); err != nil {
		logs.Errorf("sync tcloud sub account failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return err
	}

	return nil
}