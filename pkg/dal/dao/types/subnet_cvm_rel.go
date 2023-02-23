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

package types

import (
	"time"

	"hcm/pkg/dal/table/cloud"
)

// ListSubnetCvmRelDetails list subnet and cvm relation details.
type ListSubnetCvmRelDetails struct {
	Count   uint64                    `json:"count,omitempty"`
	Details []cloud.SubnetCvmRelTable `json:"details,omitempty"`
}

// ListSubnetCvmRelsJoinSubnetDetails list subnet and cvm relation joins subnet details.
type ListSubnetCvmRelsJoinSubnetDetails struct {
	Count   uint64            `json:"count,omitempty"`
	Details []SubnetWithCvmID `json:"details,omitempty"`
}

// SubnetWithCvmID subnet with cvm id.
type SubnetWithCvmID struct {
	cloud.SubnetTable `db:",inline" json:",inline"`
	CvmID             string     `db:"cvm_id" json:"cvm_id"`
	RelCreator        string     `db:"rel_creator" json:"rel_creator"`
	RelCreatedAt      *time.Time `db:"rel_created_at" json:"rel_created_at"`
}
