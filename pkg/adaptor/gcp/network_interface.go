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

package gcp

import (
	"fmt"
	"strconv"

	"hcm/pkg/adaptor/types/core"
	typesniproto "hcm/pkg/adaptor/types/network-interface"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"google.golang.org/api/compute/v1"
)

// ListNetworkInterface list network interface.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
func (g *Gcp) ListNetworkInterface(kt *kit.Kit, opt *core.GcpListOption) (*typesniproto.GcpInterfaceListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if len(opt.Zone) == 0 {
		return nil, errf.New(errf.InvalidParameter, "zone is not empty")
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID

	listCall := client.Instances.List(cloudProjectID, opt.Zone).Context(kt.Ctx)

	if opt.Page != nil {
		listCall.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := listCall.Do()
	if err != nil {
		logs.Errorf("list network interface failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	details := make([]typesniproto.GcpNI, 0, len(resp.Items))
	if err := listCall.Pages(kt.Ctx, func(page *compute.InstanceList) error {
		for _, item := range page.Items {
			for _, niItem := range item.NetworkInterfaces {
				details = append(details, converter.PtrToVal(convertNetworkInterface(item, niItem)))
			}
		}
		return nil
	}); err != nil {
		logs.Errorf("cloudapi failed to list network interface, err: %v, rid: %s", err, kt.Rid)
	}

	return &typesniproto.GcpInterfaceListResult{Details: details}, nil
}

func convertNetworkInterface(data *compute.Instance, niItem *compute.NetworkInterface) *typesniproto.GcpNI {
	v := &typesniproto.GcpNI{
		Name:          converter.ValToPtr(niItem.Name),
		Zone:          converter.ValToPtr(data.Zone),
		CloudID:       converter.ValToPtr(fmt.Sprintf("%d_%s", data.Id, niItem.Name)),
		PrivateIP:     converter.ValToPtr(niItem.NetworkIP),
		InstanceID:    converter.ValToPtr(strconv.FormatUint(data.Id, 10)),
		CloudVpcID:    converter.ValToPtr(niItem.Network),
		CloudSubnetID: converter.ValToPtr(niItem.Subnetwork),
	}

	if niItem == nil {
		return v
	}

	v.Extension = &coreni.GcpNIExtension{
		CanIpForward:  data.CanIpForward,
		Status:        data.Status,
		StackType:     niItem.StackType,
		AccessConfigs: []*coreni.AccessConfig{},
	}

	if len(niItem.AccessConfigs) != 0 {
		for _, tmpAc := range niItem.AccessConfigs {
			v.Extension.AccessConfigs = append(v.Extension.AccessConfigs, &coreni.AccessConfig{
				Type:        tmpAc.Type,
				Name:        tmpAc.Name,
				NatIP:       tmpAc.NatIP,
				NetworkTier: tmpAc.NetworkTier,
			})
		}
	}

	return v
}