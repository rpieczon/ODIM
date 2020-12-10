//(C) Copyright [2020] Hewlett Packard Enterprise Development LP
//
//Licensed under the Apache License, Version 2.0 (the "License"); you may
//not use this file except in compliance with the License. You may obtain
//a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//License for the specific language governing permissions and limitations
// under the License.

//Package rpc ...
package rpc

import (
	"context"
	"encoding/json"
	"github.com/ODIM-Project/ODIM/svc-systems/chassis"
	"log"
	"net/http"

	"github.com/ODIM-Project/ODIM/lib-utilities/common"
	chassisproto "github.com/ODIM-Project/ODIM/lib-utilities/proto/chassis"
	"github.com/ODIM-Project/ODIM/lib-utilities/response"
	"github.com/ODIM-Project/ODIM/svc-plugin-rest-client/pmbhandle"
	"github.com/ODIM-Project/ODIM/svc-systems/scommon"
)

func NewChassisRPC(
	authWrapper func(sessionToken string, privileges, oemPrivileges []string) (int32, string),
	getCollectionHandler chassis.GetCollectionHandler) *ChassisRPC {

	return &ChassisRPC{
		IsAuthorizedRPC:      authWrapper,
		GetCollectionHandler: getCollectionHandler,
	}
}

// ChassisRPC struct helps to register service
type ChassisRPC struct {
	GetCollectionHandler chassis.GetCollectionHandler
	IsAuthorizedRPC func(sessionToken string, privileges, oemPrivileges []string) response.RPC
}

func (cha *ChassisRPC) CreateChassis(_ context.Context, req *chassisproto.CreateChassisRequest, resp *chassisproto.GetChassisResponse) error {
	r := auth(cha.IsAuthorizedRPC, req.SessionToken, func() response.RPC {
		return chassis.Create(req)
	})

	return rewrite(r, resp)
}

//GetChassisResource defines the operations which handles the RPC request response
// for the getting the system resource  of systems micro service.
// The functionality retrives the request and return backs the response to
// RPC according to the protoc file defined in the util-lib package.
// The function uses IsAuthorized of util-lib to validate the session
// which is present in the request.
func (cha *ChassisRPC) GetChassisResource(ctx context.Context, req *chassisproto.GetChassisRequest, resp *chassisproto.GetChassisResponse) error {
	sessionToken := req.SessionToken
	authStatusCode, authStatusMessage := cha.IsAuthorizedRPC(sessionToken, []string{common.PrivilegeLogin}, []string{})
	if authStatusCode != http.StatusOK {
		errorMessage := "error while trying to authenticate session"
		resp.StatusCode = authStatusCode
		resp.StatusMessage = authStatusMessage
		rpcResp := common.GeneralError(authStatusCode, authStatusMessage, errorMessage, nil, nil)
		resp.Body = jsonMarshal(rpcResp.Body)
		resp.Header = rpcResp.Header
		log.Printf(errorMessage)
		return nil
	}
	var pc = chassis.PluginContact{
		ContactClient:   pmbhandle.ContactPlugin,
		DecryptPassword: common.DecryptWithPrivateKey,
		GetPluginStatus: scommon.GetPluginStatus,
	}
	data := pc.GetChassisResource(req)
	resp.Header = data.Header
	resp.StatusCode = data.StatusCode
	resp.StatusMessage = data.StatusMessage
	resp.Body = jsonMarshal(data.Body)
	return nil
}

// GetChassisCollection defines the operation which handles the RPC request response
// for getting all the server chassis added.
// Retrieves all the keys with table name ChassisCollection and create the response
// to send back to requested user.
func (cha *ChassisRPC) GetChassisCollection(_ context.Context, req *chassisproto.GetChassisRequest, resp *chassisproto.GetChassisResponse) error {
	r := auth(cha.IsAuthorizedRPC, req.SessionToken, func() response.RPC {
		return cha.GetCollectionHandler.Handle()
	})
	return rewrite(r, resp)
}

//GetChassisInfo defines the operations which handles the RPC request response
// for the getting the system resource  of systems micro service.
// The functionality retrives the request and return backs the response to
// RPC according to the protoc file defined in the util-lib package.
// The function uses IsAuthorized of util-lib to validate the session
// which is present in the request.
func (cha *ChassisRPC) GetChassisInfo(ctx context.Context, req *chassisproto.GetChassisRequest, resp *chassisproto.GetChassisResponse) error {
	sessionToken := req.SessionToken
	authStatusCode, authStatusMessage := cha.IsAuthorizedRPC(sessionToken, []string{common.PrivilegeLogin}, []string{})
	if authStatusCode != http.StatusOK {
		errorMessage := "error while trying to authenticate session"
		resp.StatusCode = authStatusCode
		resp.StatusMessage = authStatusMessage
		rpcResp := common.GeneralError(authStatusCode, authStatusMessage, errorMessage, nil, nil)
		resp.Body = jsonMarshal(rpcResp.Body)
		resp.Header = rpcResp.Header
		log.Printf(errorMessage)
		return nil
	}
	data := chassis.GetChassisInfo(req)
	resp.Header = data.Header
	resp.StatusCode = data.StatusCode
	resp.StatusMessage = data.StatusMessage
	resp.Body = jsonMarshal(data.Body)
	return nil
}

func rewrite(source response.RPC, target *chassisproto.GetChassisResponse) error {
	target.Header = source.Header
	target.StatusCode = source.StatusCode
	target.StatusMessage = source.StatusMessage
	target.Body = jsonMarshal(source.Body)
	return nil
}

func jsonMarshal(input interface{}) []byte {
	if bytes, alreadyBytes := input.([]byte); alreadyBytes {
		return bytes
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		log.Println("error in unmarshalling response object from util-libs", err.Error())
	}
	return bytes
}

func fillChassisProtoResponse(resp *chassisproto.GetChassisResponse, data response.RPC) {
	resp.StatusCode = data.StatusCode
	resp.StatusMessage = data.StatusMessage
	resp.Body = generateResponse(data.Body)
	resp.Header = data.Header
}
