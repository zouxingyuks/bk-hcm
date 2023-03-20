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

// Package vpc defines vpc service.
package vpc

import (
	"strconv"

	"hcm/cmd/hc-service/logics/subnet"
	"hcm/cmd/hc-service/service/sync"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
)

// GcpVpcCreate create gcp vpc.
func (v vpc) GcpVpcCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.VpcCreateReq[hcservice.GcpVpcCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := v.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create gcp vpc
	opt := &types.GcpVpcCreateOption{
		AccountID: req.AccountID,
		Name:      req.Name,
		Memo:      req.Memo,
		Extension: &types.GcpVpcCreateExt{
			AutoCreateSubnetworks: req.Extension.AutoCreateSubnetworks,
			EnableUlaInternalIpv6: req.Extension.EnableUlaInternalIpv6,
			InternalIpv6Range:     req.Extension.InternalIpv6Range,
			RoutingMode:           req.Extension.RoutingMode,
		},
	}
	vpcID, err := cli.CreateVpc(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	// get created vpc info
	sync.SleepBeforeSync()

	listOpt := &types.GcpListOption{CloudIDs: []string{strconv.FormatUint(vpcID, 10)}}
	listRes, err := cli.ListVpc(cts.Kit, listOpt)
	if err != nil {
		return nil, err
	}

	if len(listRes.Details) != 1 {
		return nil, errf.Newf(errf.Aborted, "get created vpc detail, but result count is invalid")
	}

	// create hcm vpc
	createReq := &cloud.VpcBatchCreateReq[cloud.GcpVpcCreateExt]{
		Vpcs: []cloud.VpcCreateReq[cloud.GcpVpcCreateExt]{convertGcpVpcCreateReq(req, &listRes.Details[0])},
	}
	result, err := v.cs.DataService().Gcp.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return nil, err
	}

	if len(result.IDs) != 1 {
		return nil, errf.New(errf.Aborted, "create result is invalid")
	}

	// create gcp subnets
	if len(req.Extension.Subnets) == 0 {
		return core.CreateResult{ID: result.IDs[0]}, nil
	}

	regionSubnetMap := make(map[string][]hcservice.SubnetCreateReq[hcservice.GcpSubnetCreateExt])
	for _, s := range req.Extension.Subnets {
		regionSubnetMap[s.Extension.Region] = append(regionSubnetMap[s.Extension.Region], s)
	}

	for region, subnets := range regionSubnetMap {
		subnetCreateOpt := &subnet.SubnetCreateOptions[hcservice.GcpSubnetCreateExt]{
			BkBizID:    req.BkBizID,
			AccountID:  req.AccountID,
			Region:     region,
			CloudVpcID: listRes.Details[0].CloudID,
			CreateReqs: subnets,
		}
		_, err = v.subnet.GcpSubnetCreate(cts.Kit, subnetCreateOpt)
		if err != nil {
			return nil, err
		}
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

func convertGcpVpcCreateReq(req *hcservice.VpcCreateReq[hcservice.GcpVpcCreateExt],
	data *types.GcpVpc) cloud.VpcCreateReq[cloud.GcpVpcCreateExt] {

	vpcReq := cloud.VpcCreateReq[cloud.GcpVpcCreateExt]{
		AccountID: req.AccountID,
		CloudID:   data.CloudID,
		BkBizID:   req.BkBizID,
		BkCloudID: req.BkCloudID,
		Name:      &data.Name,
		Region:    data.Region,
		Category:  req.Category,
		Memo:      req.Memo,
		Extension: &cloud.GcpVpcCreateExt{
			SelfLink:              data.Extension.SelfLink,
			AutoCreateSubnetworks: data.Extension.AutoCreateSubnetworks,
			EnableUlaInternalIpv6: data.Extension.EnableUlaInternalIpv6,
			InternalIpv6Range:     data.Extension.InternalIpv6Range,
			Mtu:                   data.Extension.Mtu,
			RoutingMode:           data.Extension.RoutingMode,
		},
	}

	return vpcReq
}

// GcpVpcUpdate update gcp vpc.
func (v vpc) GcpVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().Gcp.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.GcpVpcUpdateOption{
		ResourceID: getRes.CloudID,
		Data:       &types.BaseVpcUpdateData{Memo: req.Memo},
	}
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.GcpVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().Gcp.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GcpVpcDelete delete gcp vpc.
func (v vpc) GcpVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().Gcp.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.BaseDeleteOption{
		ResourceID: getRes.CloudID,
	}
	err = cli.DeleteVpc(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = v.cs.DataService().Global.Vpc.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
