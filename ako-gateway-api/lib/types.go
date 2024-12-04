/*
 * Copyright 2023-2024 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package lib

import (
	"context"
	"strings"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/google/uuid"

	"github.com/vmware/alb-sdk/go/logger"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
)

var SupportedKinds = map[gatewayv1.ProtocolType][]gatewayv1.RouteGroupKind{
	gatewayv1.HTTPProtocolType:  {{Kind: lib.HTTPRoute}},
	gatewayv1.HTTPSProtocolType: {{Kind: lib.HTTPRoute}},
}

type traceID string
type KeyContext struct {
	KeyStr string
	Ctx    context.Context
}

func getNewTraceId() string {
	traceID := uuid.New().String()
	traceID = strings.Replace(traceID, "-", "", -1) // default value
	return traceID
}

func NewKeyContextWithTraceID(key string, ctx context.Context) KeyContext {
	return KeyContext{KeyStr: key,
		Ctx: logger.SetTraceID(ctx, getNewTraceId())}
}

func NewKeyContext(key string, ctx context.Context) KeyContext {
	return KeyContext{KeyStr: key,
		Ctx: ctx}
}
