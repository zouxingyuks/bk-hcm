/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package actionlb

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// --------------------------[批量操作-批量调整RS权重]-----------------------------

var _ action.Action = new(BatchTaskModifyRsWeightAction)
var _ action.ParameterAction = new(BatchTaskModifyRsWeightAction)

// BatchTaskModifyRsWeightAction 批量操作-调整RS权重
type BatchTaskModifyRsWeightAction struct{}

// BatchTaskModifyRsWeightOption ...
type BatchTaskModifyRsWeightOption struct {
	Vendor         enumor.Vendor `json:"vendor" validate:"required"`
	LoadBalancerID string        `json:"lb_id" validate:"required"`
	// ManagementDetailIDs 对应的详情行id列表，需要和批量绑定的Targets参数长度对应
	ManagementDetailIDs []string                             `json:"management_detail_ids" validate:"required,max=20"`
	LblList             []*hclb.TCloudBatchModifyRsWeightReq `json:"lbl_list" validate:"required,max=20"`
}

// Validate validate option.
func (opt BatchTaskModifyRsWeightOption) Validate() error {
	if opt.LblList == nil {
		return errf.New(errf.InvalidParameter, "lbl_list is required")
	}
	if len(opt.ManagementDetailIDs) != len(opt.LblList) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and lbl list num not match, %d != %d",
			len(opt.ManagementDetailIDs), len(opt.LblList))
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act BatchTaskModifyRsWeightAction) ParameterNew() (params any) {
	return new(BatchTaskModifyRsWeightOption)
}

// Name return action name
func (act BatchTaskModifyRsWeightAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskTCloudModifyRsWeight
}

// Run 批量解绑RS
func (act BatchTaskModifyRsWeightAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskModifyRsWeightOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter,
			fmt.Sprintf("params type mismatch, BatchTaskModifyRsWeightOption:%+v", params))
	}

	results := make([]*hclb.BatchCreateResult, 0, len(opt.LblList))
	for i := range opt.LblList {
		detailID := opt.ManagementDetailIDs[i]
		// 逐条更新结果
		ret, optErr := act.batchListenerModifyRsWeight(kt.Kit(), opt.LoadBalancerID, detailID, opt.LblList[i])
		// 结束后写回状态
		targetState := enumor.TaskDetailSuccess
		if optErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err := batchUpdateTaskDetailResultState(kt.Kit(), []string{detailID}, targetState, ret, optErr)
		if err != nil {
			logs.Errorf("failed to set detail to [%s] after cloud operation finished, err: %v, rid: %s",
				targetState, err, kt.Kit().Rid)
			return nil, err
		}
		if optErr != nil {
			// abort
			return nil, err
		}
		results = append(results, ret)
	}
	// all success
	return results, nil
}

// batchListenerModifyRsWeight 批量调整监听器RS的权重
func (act BatchTaskModifyRsWeightAction) batchListenerModifyRsWeight(kt *kit.Kit, lbID, detailID string,
	req *hclb.TCloudBatchModifyRsWeightReq) (*hclb.BatchCreateResult, error) {

	detailList, err := listTaskDetail(kt, []string{detailID})
	if err != nil {
		logs.Errorf("failed to query task detail, err: %v, detailID: %s, rid: %s", err, detailID, kt.Rid)
		return nil, err
	}

	detail := detailList[0]
	if detail.State == enumor.TaskDetailCancel {
		// 任务被取消，跳过该任务, 直接成功即可
		return nil, nil
	}
	if detail.State != enumor.TaskDetailInit {
		return nil, errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
			detail.ID, detail.State)
	}

	// 更新任务状态为 running
	if err = batchUpdateTaskDetailState(kt, []string{detailID}, enumor.TaskDetailRunning); err != nil {
		return nil, fmt.Errorf("failed to update detail to running, detailID: %s, err: %v", detailID, err)
	}

	// 调用云API调整RS权重，支持幂等
	lblResp := &hclb.BatchCreateResult{}
	switch req.Vendor {
	case enumor.TCloud:
		lblResp, err = actcli.GetHCService().TCloud.Clb.BatchModifyListenerTargetsWeight(kt, lbID, req)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "batch listener modify rs weight failed, invalid vendor: %s",
			req.Vendor)
	}
	if err != nil {
		logs.Errorf("failed to call hc to listener modify rs weight, err: %v, lbID: %s, detailID: %s, rid: %s",
			err, lbID, detailID, kt.Rid)
		return nil, err
	}
	return lblResp, nil
}

// Rollback 调整RS权重支持重入，无需回滚
func (act BatchTaskModifyRsWeightAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskModifyRsWeightAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}