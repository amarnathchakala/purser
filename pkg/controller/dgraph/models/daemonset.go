/*
 * Copyright (c) 2018 VMware Inc. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package models

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vmware/purser/pkg/controller/dgraph"
	ext_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// Dgraph Model Constants
const (
	IsDaemonset = "isDaemonset"
)

// Daemonset schema in dgraph
type Daemonset struct {
	dgraph.ID
	IsDaemonset bool       `json:"isDaemonset,omitempty"`
	Name        string     `json:"name,omitempty"`
	StartTime   string     `json:"startTime,omitempty"`
	EndTime     string     `json:"endTime,omitempty"`
	Namespace   *Namespace `json:"namespace,omitempty"`
	Pods        []*Pod     `json:"pod,omitempty"`
	Type        string     `json:"type,omitempty"`
}

func createDaemonsetObject(daemonset ext_v1beta1.DaemonSet) Daemonset {
	newDaemonset := Daemonset{
		Name:        "daemonset-" + daemonset.Name,
		IsDaemonset: true,
		Type:        "daemonset",
		ID:          dgraph.ID{Xid: daemonset.Namespace + ":" + daemonset.Name},
		StartTime:   daemonset.GetCreationTimestamp().Time.Format(time.RFC3339),
	}
	namespaceUID := CreateOrGetNamespaceByID(daemonset.Namespace)
	if namespaceUID != "" {
		newDaemonset.Namespace = &Namespace{ID: dgraph.ID{UID: namespaceUID, Xid: daemonset.Namespace}}
	}
	daemonsetDeletionTimestamp := daemonset.GetDeletionTimestamp()
	if !daemonsetDeletionTimestamp.IsZero() {
		newDaemonset.EndTime = daemonsetDeletionTimestamp.Time.Format(time.RFC3339)
		newDaemonset.Xid += newDaemonset.EndTime
		newDaemonset.Name += "*" + newDaemonset.EndTime
	}
	return newDaemonset
}

// StoreDaemonset create a new daemonset in the Dgraph and updates if already present.
func StoreDaemonset(daemonset ext_v1beta1.DaemonSet) (string, error) {
	xid := daemonset.Namespace + ":" + daemonset.Name
	uid := dgraph.GetUID(xid, IsDaemonset)

	newDaemonset := createDaemonsetObject(daemonset)
	if uid != "" {
		newDaemonset.UID = uid
	}
	assigned, err := dgraph.MutateNode(newDaemonset, dgraph.CREATE)
	if err != nil {
		return "", err
	}
	return assigned.Uids["blank-0"], nil
}

// CreateOrGetDaemonsetByID returns the uid of namespace if exists,
// otherwise creates the daemonset and returns uid.
func CreateOrGetDaemonsetByID(xid string) string {
	if xid == "" {
		return ""
	}
	uid := dgraph.GetUID(xid, IsDaemonset)

	if uid != "" {
		return uid
	}

	d := Daemonset{
		ID:          dgraph.ID{Xid: xid},
		Name:        xid,
		IsDaemonset: true,
	}
	assigned, err := dgraph.MutateNode(d, dgraph.CREATE)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return assigned.Uids["blank-0"]
}
