/*
Copyright 2023 VMware, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"context"
	"fmt"

	logr "github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

func createGatewayClass(gwApiClient *gatewayclientset.Clientset, log logr.Logger) error {
	var err error
	gwClassOnce.Do(func() {
		gatewayClass, err = readGatewayClassFromManifest(gatewayClassLocation, log)
	})
	existingGWClass, err := gwApiClient.GatewayV1beta1().GatewayClasses().Get(context.TODO(), gatewayClassName, v1.GetOptions{})
	if err == nil && existingGWClass != nil {
		if gatewayClass.GetResourceVersion() == existingGWClass.GetResourceVersion() {
			log.Info(fmt.Sprintf("no updates required for %s Gateway Class", gatewayClassName))
		} else {
			gatewayClass.SetResourceVersion(existingGWClass.GetResourceVersion())
			_, err = gwApiClient.GatewayV1beta1().GatewayClasses().Update(context.TODO(), gatewayClass, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s Gateway Class", gatewayClassName))
				return err
			} else {
				log.Info(fmt.Sprintf("successfully updated %s Gateway Class", gatewayClassName))
			}
		}
		return nil
	}

	_, err = gwApiClient.GatewayV1beta1().GatewayClasses().Create(context.TODO(), gatewayClass, v1.CreateOptions{})
	if err == nil {
		log.Info(fmt.Sprintf("%s Gateway Class created", gatewayClassName))
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.Info(fmt.Sprintf("%s Gateway Class already exists", gatewayClassName))
		return nil
	}
	return err
}
