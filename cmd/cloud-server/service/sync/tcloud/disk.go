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

	"hcm/pkg/api/hc-service/disk"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncDisk ...
func SyncDisk(kt *kit.Kit, service *hcservice.Client, accountID string, regions []string) error {

	start := time.Now()
	logs.V(3).Infof("tcloud account[%s] sync disk start, time: %v, rid: %s", accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("tcloud account[%s] sync disk end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
	}()

	for _, region := range regions {
		req := &disk.DiskSyncReq{
			AccountID: accountID,
			Region:    region,
		}
		if err := service.TCloud.Disk.SyncDisk(kt.Ctx, kt.Header(), req); err != nil {
			logs.Errorf("sync tcloud disk failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
			return err
		}
	}

	return nil
}