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

//Package managers ...
package managers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	dmtf "github.com/ODIM-Project/ODIM/lib-dmtf/model"
	"github.com/ODIM-Project/ODIM/lib-utilities/common"
	"github.com/ODIM-Project/ODIM/lib-utilities/config"
	"github.com/ODIM-Project/ODIM/lib-utilities/errors"
	managersproto "github.com/ODIM-Project/ODIM/lib-utilities/proto/managers"
	"github.com/ODIM-Project/ODIM/lib-utilities/response"
	"github.com/ODIM-Project/ODIM/svc-managers/mgrcommon"
	"github.com/ODIM-Project/ODIM/svc-managers/mgrmodel"
	"github.com/ODIM-Project/ODIM/svc-managers/mgrresponse"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

var (
	JsonUnMarshalFunc              = json.Unmarshal
	RequestParamsCaseValidatorFunc = common.RequestParamsCaseValidator
)

// GetManagersCollection will get the all the managers(odimra, Plugins, Servers)
func (e *ExternalInterface) GetManagersCollection(req *managersproto.ManagerRequest) (response.RPC, error) {
	var resp response.RPC
	managers := mgrresponse.ManagersCollection{
		OdataContext: "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
		OdataID:      "/redfish/v1/Managers",
		OdataType:    "#ManagerCollection.ManagerCollection",
		Description:  "Managers view",
		Name:         "Managers",
	}
	var members []dmtf.Link

	// Add servers as manager in manager collection
	managersCollectionKeysArray, err := e.DB.GetAllKeysFromTable("Managers")
	if err != nil || len(managersCollectionKeysArray) == 0 {
		log.Error("odimra Doesnt have Servers")
	}

	for _, key := range managersCollectionKeysArray {
		members = append(members, dmtf.Link{Oid: key})
	}
	managers.Members = members
	managers.MembersCount = len(members)
	resp.Body = managers
	resp.StatusCode = http.StatusOK
	return resp, nil
}

// GetManagers will fetch individual manager details with the given ID
func (e *ExternalInterface) GetManagers(req *managersproto.ManagerRequest) response.RPC {
	var resp response.RPC
	if req.ManagerID == config.Data.RootServiceUUID {
		manager, err := e.getManagerDetails(req.ManagerID)
		if err != nil {
			log.Error("error getting manager details : " + err.Error())
			errArgs := []interface{}{"Managers", req.ManagerID}
			errorMessage := err.Error()
			resp = common.GeneralError(http.StatusNotFound, response.ResourceNotFound, errorMessage,
				errArgs, nil)
			return resp
		}
		resp.Body = manager
	} else {
		requestData := strings.SplitN(req.ManagerID, ".", 2)
		if len(requestData) <= 1 {
			resp = e.getPluginManagerResoure(requestData[0], req.URL)
			return resp
		}
		uuid := requestData[0]
		data, err := e.DB.GetManagerByURL(req.URL)
		if err != nil {
			log.Error("error getting manager details : " + err.Error())
			var errArgs = []interface{}{}
			var statusCode int
			var StatusMessage string
			errorMessage := err.Error()
			if errors.DBKeyNotFound == err.ErrNo() {
				errArgs = []interface{}{"Managers", req.ManagerID}

				statusCode = http.StatusNotFound
				StatusMessage = response.ResourceNotFound
			} else {
				statusCode = http.StatusInternalServerError
				StatusMessage = response.InternalError
			}
			resp = common.GeneralError(int32(statusCode), StatusMessage, errorMessage,
				errArgs, nil)
			return resp
		}
		var managerData map[string]interface{}
		jerr := json.Unmarshal([]byte(data), &managerData)
		if jerr != nil {
			errorMessage := "error unmarshalling manager details: " + jerr.Error()
			log.Error(errorMessage)
			resp = common.GeneralError(http.StatusInternalServerError, response.InternalError, errorMessage,
				nil, nil)
			return resp
		}
		// extracting the Manager Type from the  managerData
		var managerType string
		if val, ok := managerData["ManagerType"]; ok {
			managerType = val.(string)
		}
		//adding default description
		if _, ok := managerData["Description"]; !ok {
			managerData["Description"] = "BMC Manager"
		}
		//adding RemoteAccountService object to manager response
		if _, ok := managerData["RemoteAccountService"]; !ok {
			managerData["RemoteAccountService"] = map[string]string{
				"@odata.id": "/redfish/v1/Managers/" + req.ManagerID + "/RemoteAccountService",
			}
		}
		//adding PowerState
		if _, ok := managerData["PowerState"]; !ok {
			managerData["PowerState"] = "On"
		}
		if managerType != common.ManagerTypeService && managerType != "" {
			deviceData, err := e.getResourceInfoFromDevice(req.URL, uuid, requestData[1])
			if err != nil {
				log.Error("Device " + req.URL + " is unreachable: " + err.Error())
				// Updating the state
				managerData["Status"] = map[string]string{
					"State": "Absent",
				}
			} else {
				jerr := json.Unmarshal([]byte(deviceData), &managerData)
				if jerr != nil {
					errorMessage := "error unmarshaling manager details: " + jerr.Error()
					log.Error(errorMessage)
					resp = common.GeneralError(http.StatusInternalServerError, response.InternalError, errorMessage,
						nil, nil)
					return resp
				}
			}
			err = e.DB.UpdateData(req.URL, managerData, "Managers")
			if err != nil {
				errorMessage := "error while saving manager details: " + err.Error()
				log.Error(errorMessage)
				resp = common.GeneralError(http.StatusInternalServerError, response.InternalError, errorMessage,
					nil, nil)
				return resp
			}
			dataBytes, err := json.Marshal(managerData)
			if err != nil {
				errorMessage := "error while marshalling manager details: " + err.Error()
				log.Error(errorMessage)
				resp = common.GeneralError(http.StatusInternalServerError, response.InternalError, errorMessage,
					nil, nil)
				return resp
			}
			data = string(dataBytes)
		}
		var resource map[string]interface{}
		json.Unmarshal([]byte(data), &resource)
		resp.Body = resource
	}
	resp.StatusCode = http.StatusOK
	resp.StatusMessage = response.Success
	return resp
}

func (e *ExternalInterface) getManagerDetails(id string) (mgrmodel.Manager, error) {
	var mgr mgrmodel.Manager
	var mgrData mgrmodel.RAManager
	data, err := e.DB.GetManagerByURL("/redfish/v1/Managers/" + id)
	if err != nil {
		return mgr, fmt.Errorf("unable to retrieve manager information: %v", err)
	}

	if err := JsonUnMarshalFunc([]byte(data), &mgrData); err != nil {
		return mgr, fmt.Errorf("unable to marshal manager information: %v", err)
	}

	chassisList, chassisErr := e.DB.GetAllKeysFromTable("Chassis")
	if chassisErr != nil {
		return mgr, fmt.Errorf("unable to retrieve chassis list information: %v", chassisErr)
	}

	serverList, serverErr := e.DB.GetAllKeysFromTable("ComputerSystem")
	if serverErr != nil {
		return mgr, fmt.Errorf("unable to retrieve server list information: %v", serverErr)
	}
	managerList, mgrErr := e.DB.GetAllKeysFromTable("Managers")
	if mgrErr != nil {
		return mgr, fmt.Errorf("unable to retrieve manager list information: %v", mgrErr)
	}
	var chassisLink, serverLink, managerLink []*dmtf.Link
	if len(chassisList) > 0 {
		for _, key := range chassisList {
			chassisLink = append(chassisLink, &dmtf.Link{Oid: key})
		}
	}
	if len(serverList) > 0 {
		for _, key := range serverList {
			serverLink = append(serverLink, &dmtf.Link{Oid: key})
		}
	}
	odimURI := "/redfish/v1/Managers/" + config.Data.RootServiceUUID
	if len(managerList) > 0 {
		for _, key := range managerList {
			if key != odimURI {
				managerLink = append(managerLink, &dmtf.Link{Oid: key})
			}
		}
	}

	return mgrmodel.Manager{
		OdataContext:    "/redfish/v1/$metadata#Manager.Manager",
		OdataID:         "/redfish/v1/Managers/" + id,
		OdataType:       common.ManagerType,
		Name:            mgrData.Name,
		ManagerType:     mgrData.ManagerType,
		ID:              mgrData.ID,
		UUID:            mgrData.UUID,
		FirmwareVersion: mgrData.FirmwareVersion,
		Status: &mgrmodel.Status{
			State:  mgrData.State,
			Health: mgrData.Health,
		},
		Links: &mgrmodel.Links{
			ManagerForChassis:  chassisLink,
			ManagerForServers:  serverLink,
			ManagerForManagers: managerLink,
		},
		Description:         mgrData.Description,
		LogServices:         mgrData.LogServices,
		Model:               mgrData.Model,
		DateTime:            time.Now().Format(time.RFC3339),
		DateTimeLocalOffset: "+00:00",
		PowerState:          mgrData.PowerState,
	}, nil
}

// GetManagersResource is used to fetch resource data. The function is supposed to be used as part of RPC
// For getting system resource information,  parameters need to be passed GetSystemsRequest .
// GetManagersResource holds the  Uuid,Url and Resourceid ,
// Url will be parsed from that search key will created
// There will be two return values for the fuction. One is the RPC response, which contains the
// status code, status message, headers and body and the second value is error.
func (e *ExternalInterface) GetManagersResource(req *managersproto.ManagerRequest) response.RPC {
	var resp response.RPC
	var tableName string
	var resourceName string
	var resource map[string]interface{}
	requestData := strings.SplitN(req.ManagerID, ".", 2)
	urlData := strings.Split(req.URL, "/")
	if len(requestData) <= 1 {
		if req.ResourceID == "" {
			resourceName = urlData[len(urlData)-1]
			tableName = common.ManagersResource[resourceName]
		} else {
			tableName = urlData[len(urlData)-2]
		}
		data, err := e.DB.GetResource(tableName, req.URL)
		if err != nil {
			if req.ManagerID != config.Data.RootServiceUUID {
				return e.getPluginManagerResoure(requestData[0], req.URL)
			}
			errorMessage := "unable to get odimra managers details: " + err.Error()
			log.Error(errorMessage)
			return common.GeneralError(http.StatusInternalServerError, response.InternalError, errorMessage, []interface{}{}, nil)
		}

		json.Unmarshal([]byte(data), &resource)
		resp.Body = resource
		resp.StatusCode = http.StatusOK
		resp.StatusMessage = response.Success

		return resp

	}
	uuid := requestData[0]

	if req.ResourceID == "" {
		resourceName := urlData[len(urlData)-1]
		tableName = common.ManagersResource[resourceName]
	} else {
		tableName = urlData[len(urlData)-2]
	}
	data, err := e.DB.GetResource(tableName, req.URL)
	if err != nil {
		if errors.DBKeyNotFound == err.ErrNo() {
			var err error
			if data, err = e.getResourceInfoFromDevice(req.URL, uuid, requestData[1]); err != nil {
				errorMessage := "unable to get resource details from device: " + err.Error()
				log.Error(errorMessage)
				errArgs := []interface{}{tableName, req.ManagerID}
				return common.GeneralError(http.StatusNotFound, response.ResourceNotFound, errorMessage, errArgs, nil)
			}
		} else {
			errorMessage := "unable to get managers details: " + err.Error()
			log.Error(errorMessage)
			return common.GeneralError(http.StatusInternalServerError, response.InternalError, errorMessage, []interface{}{}, nil)
		}
	}

	json.Unmarshal([]byte(data), &resource)

	if common.Types[tableName] != "" && resource != nil {
		resource["@odata.type"] = common.Types[tableName]
	}

	resp.Body = resource
	resp.StatusCode = http.StatusOK
	resp.StatusMessage = response.Success

	return resp
}

// VirtualMediaActions is used to perform action on VirtualMedia. For insert and eject of virtual media this function is used
func (e *ExternalInterface) VirtualMediaActions(req *managersproto.ManagerRequest) response.RPC {
	var resp response.RPC
	var requestBody = req.RequestBody
	//InsertMedia payload validation
	if strings.Contains(req.URL, "VirtualMedia.InsertMedia") {
		var vmiReq mgrmodel.VirtualMediaInsert
		// Updating the default values
		vmiReq.Inserted = true
		vmiReq.WriteProtected = true
		err := json.Unmarshal(req.RequestBody, &vmiReq)
		if err != nil {
			errorMessage := "while unmarshaling the virtual media insert request: " + err.Error()
			log.Error(errorMessage)
			resp = common.GeneralError(http.StatusBadRequest, response.MalformedJSON, errorMessage, []interface{}{}, nil)
			return resp
		}

		// Validating the request JSON properties for case sensitive
		invalidProperties, err := RequestParamsCaseValidatorFunc(req.RequestBody, vmiReq)
		if err != nil {
			errMsg := "while validating request parameters for virtual media insert: " + err.Error()
			log.Error(errMsg)
			return common.GeneralError(http.StatusInternalServerError, response.InternalError, errMsg, nil, nil)
		} else if invalidProperties != "" {
			errorMessage := "one or more properties given in the request body are not valid, ensure properties are listed in uppercamelcase "
			log.Error(errorMessage)
			response := common.GeneralError(http.StatusBadRequest, response.PropertyUnknown, errorMessage, []interface{}{invalidProperties}, nil)
			return response
		}

		// Check mandatory fields
		statuscode, statusMessage, messageArgs, err := validateFields(&vmiReq)
		if err != nil {
			errorMessage := "request payload validation failed: " + err.Error()
			log.Error(errorMessage)
			resp = common.GeneralError(statuscode, statusMessage, errorMessage, messageArgs, nil)
			return resp
		}
		requestBody, err = json.Marshal(vmiReq)
		if err != nil {
			log.Error("while marshalling the virtual media insert request: " + err.Error())
			resp = common.GeneralError(http.StatusInternalServerError, response.InternalError, err.Error(), nil, nil)
			return resp
		}
	}
	// splitting managerID to get uuid
	requestData := strings.SplitN(req.ManagerID, ".", 2)
	uuid := requestData[0]
	resp = e.deviceCommunication(req.URL, uuid, requestData[1], http.MethodPost, requestBody)

	// If the virtualmedia action is success then updating DB
	if resp.StatusCode == http.StatusOK {
		vmURI := strings.Replace(req.URL, "/Actions/VirtualMedia.InsertMedia", "", -1)
		vmURI = strings.Replace(vmURI, "/Actions/VirtualMedia.EjectMedia", "", -1)
		deviceData, err := e.getResourceInfoFromDevice(vmURI, uuid, requestData[1])
		if err != nil {
			log.Error("while trying get on URI " + vmURI + " : " + err.Error())
		} else {
			var vmData map[string]interface{}
			jerr := json.Unmarshal([]byte(deviceData), &vmData)
			if jerr != nil {
				log.Error("while unmarshaling virtual media details: " + jerr.Error())
			} else {
				err = e.DB.UpdateData(vmURI, vmData, "VirtualMedia")
				if err != nil {
					log.Error("while saving virtual media details: " + err.Error())
				}
			}
		}
	}
	return resp
}

// validateFields will validate the request payload, if any mandatory fields are missing then it will generate an error
func validateFields(request *mgrmodel.VirtualMediaInsert) (int32, string, []interface{}, error) {
	validate := validator.New()
	// if any of the mandatory fields missing in the struct, then it will return an error
	err := validate.Struct(request)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return http.StatusBadRequest, response.PropertyMissing, []interface{}{err.Field()}, fmt.Errorf(err.Field() + " field is missing")
		}
	}
	return http.StatusOK, common.OK, []interface{}{}, nil
}

func (e *ExternalInterface) getPluginManagerResoure(managerID, reqURI string) response.RPC {
	var resp response.RPC
	data, dberr := e.DB.GetManagerByURL("/redfish/v1/Managers/" + managerID)
	if dberr != nil {
		log.Error("unable to get manager details : " + dberr.Error())
		var errArgs = []interface{}{"Managers", managerID}
		errorMessage := dberr.Error()
		resp = common.GeneralError(http.StatusNotFound, response.ResourceNotFound, errorMessage,
			errArgs, nil)
		return resp
	}
	var managerData map[string]interface{}
	jerr := json.Unmarshal([]byte(data), &managerData)
	if jerr != nil {
		errorMessage := "unable to unmarshal manager details: " + jerr.Error()
		log.Error(errorMessage)
		resp = common.GeneralError(http.StatusInternalServerError, response.InternalError, errorMessage, nil, nil)
		return resp
	}
	var pluginID = managerData["Name"].(string)
	// Get the Plugin info
	plugin, gerr := e.DB.GetPluginData(pluginID)
	if gerr != nil {
		log.Error("unable to get manager details : " + gerr.Error())
		var errArgs = []interface{}{"Plugin", pluginID}
		errorMessage := gerr.Error()
		resp = common.GeneralError(http.StatusNotFound, response.ResourceNotFound, errorMessage,
			errArgs, nil)
		return resp
	}
	var req mgrcommon.PluginContactRequest

	req.ContactClient = e.Device.ContactClient
	req.Plugin = plugin

	if strings.EqualFold(plugin.PreferredAuthType, "XAuthToken") {
		token := mgrcommon.GetPluginToken(req)
		if token == "" {
			var errorMessage = "unable to create session with plugin " + plugin.ID
			return common.GeneralError(http.StatusUnauthorized, response.NoValidSession, errorMessage,
				[]interface{}{}, nil)
		}
		req.Token = token
	} else {
		req.BasicAuth = map[string]string{
			"UserName": plugin.Username,
			"Password": string(plugin.Password),
		}

	}

	req.OID = reqURI
	var errorMessage = "unable to get the details " + reqURI + ": "
	body, _, getResponse, err := mgrcommon.ContactPlugin(req, errorMessage)
	if err != nil {
		if getResponse.StatusCode == http.StatusUnauthorized && strings.EqualFold(req.Plugin.PreferredAuthType, "XAuthToken") {
			if body, _, getResponse, err = mgrcommon.RetryManagersOperation(req, errorMessage); err != nil {
				resp.StatusCode = getResponse.StatusCode
				json.Unmarshal(body, &resp.Body)
				return resp
			}
		} else {
			resp.StatusCode = getResponse.StatusCode
			json.Unmarshal(body, &resp.Body)
			return resp
		}
	}

	return fillResponse(body, managerData)

}

func fillResponse(body []byte, managerData map[string]interface{}) response.RPC {
	var resp response.RPC
	data := string(body)
	//replacing the response with north bound translation URL
	for key, value := range config.Data.URLTranslation.NorthBoundURL {
		data = strings.Replace(data, key, value, -1)
	}
	var respData map[string]interface{}
	err := json.Unmarshal([]byte(data), &respData)
	if err != nil {
		log.Error(err.Error())
		return common.GeneralError(http.StatusInternalServerError, response.InternalError, err.Error(),
			[]interface{}{}, nil)
	}
	//To populate current Datetime and DateTimeLocalOffset for Plugin manager
	respData["DateTime"] = time.Now().Format(time.RFC3339)
	respData["DateTimeLocalOffset"] = "+00:00"

	if _, ok := respData["SerialConsole"]; !ok {
		respData["SerialConsole"] = dmtf.SerialConsole{}
	}
	respData["Links"] = managerData["Links"]
	resp.Body = respData
	resp.StatusCode = http.StatusOK
	resp.StatusMessage = response.Success
	return resp

}

func (e *ExternalInterface) getResourceInfoFromDevice(reqURL, uuid, systemID string) (string, error) {
	var getDeviceInfoRequest = mgrcommon.ResourceInfoRequest{
		URL:                   reqURL,
		UUID:                  uuid,
		SystemID:              systemID,
		ContactClient:         e.Device.ContactClient,
		DecryptDevicePassword: e.Device.DecryptDevicePassword,
	}
	return e.Device.GetDeviceInfo(getDeviceInfoRequest)

}

func (e *ExternalInterface) deviceCommunication(reqURL, uuid, systemID, httpMethod string, requestBody []byte) response.RPC {
	var deviceInfoRequest = mgrcommon.ResourceInfoRequest{
		URL:                   reqURL,
		UUID:                  uuid,
		SystemID:              systemID,
		ContactClient:         e.Device.ContactClient,
		DecryptDevicePassword: e.Device.DecryptDevicePassword,
		HTTPMethod:            httpMethod,
		RequestBody:           requestBody,
	}
	return e.Device.DeviceRequest(deviceInfoRequest)
}

// GetRemoteAccountService is used to fetch resource data for BMC account service.
// ManagerRequest holds the UUID, URL and ResourceId ,
// There will be two return values for the function. One is the RPC response, which contains the
// status code, status message, headers and body and the second value is error.
func (e *ExternalInterface) GetRemoteAccountService(req *managersproto.ManagerRequest) response.RPC {
	var resp response.RPC

	requestData := strings.SplitN(req.ManagerID, ".", 2)
	uuid := requestData[0]
	uri := replaceBMCAccReq(req.URL, req.ManagerID)
	data, err := e.getResourceInfoFromDevice(uri, uuid, requestData[1])
	if err != nil {
		return handleRemoteAccountServiceError(req.URL, req.ManagerID, err)
	}
	// Replace response body to BMC manager
	data = replaceBMCAccResp(data, req.ManagerID)
	resource := convertToRedfishModel(req.URL, data)
	resp.Body = resource
	resp.StatusCode = http.StatusOK
	resp.StatusMessage = response.Success
	return resp
}

func handleRemoteAccountServiceError(uri, managerID string, err error) response.RPC {
	errorMessage := "unable to get resource details from device: " + err.Error()
	log.Error(errorMessage)
	URIRegexAcc := regexp.MustCompile(`^\/redfish\/v1\/Managers\/[a-zA-Z0-9._-]+\/RemoteAccountService\/Accounts\/[a-zA-Z0-9._-]+[\/]?$`)
	URIRegexRoles := regexp.MustCompile(`^\/redfish\/v1\/Managers\/[a-zA-Z0-9._-]+\/RemoteAccountService\/Roles\/[a-zA-Z0-9._-]+[\/]?$`)
	if URIRegexAcc.MatchString(uri) {
		accID := uri[strings.LastIndex(uri, "/")+1:]
		errArgs := []interface{}{"Accounts", accID}
		return common.GeneralError(http.StatusNotFound, response.ResourceNotFound, errorMessage, errArgs, nil)
	} else if URIRegexRoles.MatchString(uri) {
		roleID := uri[strings.LastIndex(uri, "/")+1:]
		errArgs := []interface{}{"Roles", roleID}
		return common.GeneralError(http.StatusNotFound, response.ResourceNotFound, errorMessage, errArgs, nil)
	}
	errArgs := []interface{}{"Managers", managerID}
	return common.GeneralError(http.StatusNotFound, response.ResourceNotFound, errorMessage, errArgs, nil)
}

func convertToRedfishModel(uri, data string) interface{} {
	URIRegexRemAcc := regexp.MustCompile(`^\/redfish\/v1\/Managers\/[a-zA-Z0-9._-]+\/RemoteAccountService+[\/]?$`)
	URIRegexAcc := regexp.MustCompile(`^\/redfish\/v1\/Managers\/[a-zA-Z0-9._-]+\/RemoteAccountService\/Accounts\/[a-zA-Z0-9._-]+[\/]?$`)
	URIRegexRoles := regexp.MustCompile(`^\/redfish\/v1\/Managers\/[a-zA-Z0-9._-]+\/RemoteAccountService\/Roles\/[a-zA-Z0-9._-]+[\/]?$`)
	if URIRegexRemAcc.MatchString(uri) {
		var resource dmtf.AccountService
		json.Unmarshal([]byte(data), &resource)
		return resource
	} else if URIRegexAcc.MatchString(uri) {
		var resource dmtf.ManagerAccount
		json.Unmarshal([]byte(data), &resource)
		return resource
	} else if URIRegexRoles.MatchString(uri) {
		var resource dmtf.Role
		json.Unmarshal([]byte(data), &resource)
		return resource
	}
	var resource map[string]interface{}
	json.Unmarshal([]byte(data), &resource)
	return resource
}

// CreateRemoteAccountService is used to create BMC account user
func (e *ExternalInterface) CreateRemoteAccountService(req *managersproto.ManagerRequest) response.RPC {
	var resp response.RPC
	var requestBody = req.RequestBody
	var bmcAccReq mgrmodel.CreateBMCAccount
	// Updating the default values
	err := json.Unmarshal(req.RequestBody, &bmcAccReq)
	if err != nil {
		errorMessage := "while unmarshaling the create remote account service request: " + err.Error()
		log.Error(errorMessage)
		resp = common.GeneralError(http.StatusBadRequest, response.MalformedJSON, errorMessage, []interface{}{}, nil)
		return resp
	}

	// Validating the request JSON properties for case sensitive
	invalidProperties, err := RequestParamsCaseValidatorFunc(req.RequestBody, bmcAccReq)
	if err != nil {
		errMsg := "while validating request parameters for creating BMC account: " + err.Error()
		log.Error(errMsg)
		return common.GeneralError(http.StatusInternalServerError, response.InternalError, errMsg, nil, nil)
	} else if invalidProperties != "" {
		errorMessage := "one or more properties given in the request body are not valid, ensure properties are listed in uppercamelcase "
		log.Error(errorMessage)
		response := common.GeneralError(http.StatusBadRequest, response.PropertyUnknown, errorMessage, []interface{}{invalidProperties}, nil)
		return response
	}

	// Check mandatory fields
	statuscode, statusMessage, messageArgs, err := validateCreateRemoteAccFields(&bmcAccReq)
	if err != nil {
		errorMessage := "request payload validation failed: " + err.Error()
		log.Error(errorMessage)
		resp = common.GeneralError(statuscode, statusMessage, errorMessage, messageArgs, nil)
		return resp
	}
	requestBody, err = json.Marshal(bmcAccReq)
	if err != nil {
		log.Error("while marshalling the create BMC account request: " + err.Error())
		resp = common.GeneralError(http.StatusInternalServerError, response.InternalError, err.Error(), nil, nil)
		return resp
	}
	// splitting managerID to get uuid
	requestData := strings.SplitN(req.ManagerID, ".", 2)
	uuid := requestData[0]

	uri := replaceBMCAccReq(req.URL, req.ManagerID)
	resp = e.deviceCommunication(uri, uuid, requestData[1], http.MethodPost, requestBody)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		body, _ := json.Marshal(resp.Body)
		respBody := replaceBMCAccResp(string(body), req.ManagerID)
		var managerAcc dmtf.ManagerAccount
		json.Unmarshal([]byte(respBody), &managerAcc)
		resp.Body = managerAcc
		resp.StatusCode = http.StatusCreated
	}
	return resp
}

// validateFields will validate the request payload, if any mandatory fields are missing then it will generate an error
func validateCreateRemoteAccFields(request *mgrmodel.CreateBMCAccount) (int32, string, []interface{}, error) {
	validate := validator.New()
	// if any of the mandatory fields missing in the struct, then it will return an error
	err := validate.Struct(request)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return http.StatusBadRequest, response.PropertyMissing, []interface{}{err.Field()}, fmt.Errorf(err.Field() + " field is missing")
		}
	}
	return http.StatusOK, common.OK, []interface{}{}, nil
}

func replaceBMCAccReq(uri, managerID string) string {
	uri = strings.Replace(uri, "Managers/"+managerID+"/Remote", "", -1)
	return uri
}

func replaceBMCAccResp(data, managerID string) string {
	data = strings.Replace(data, "v1/AccountService", "v1/Managers/"+managerID+"/RemoteAccountService", -1)
	return data
}

// UpdateRemoteAccountService is used to update BMC account
func (e *ExternalInterface) UpdateRemoteAccountService(req *managersproto.ManagerRequest) response.RPC {
	var resp response.RPC
	var bmcAccReq mgrmodel.UpdateBMCAccount

	// Updating the default values
	err := json.Unmarshal(req.RequestBody, &bmcAccReq)
	if err != nil {
		errorMessage := "while unmarshaling the update remote account service request: " + err.Error()
		log.Error(errorMessage)
		resp = common.GeneralError(http.StatusBadRequest, response.MalformedJSON, errorMessage, []interface{}{}, nil)
		return resp
	}

	// Validating the request JSON properties for case sensitive
	invalidProperties, err := RequestParamsCaseValidatorFunc(req.RequestBody, bmcAccReq)
	if err != nil {
		errMsg := "while validating request parameters for updating BMC account: " + err.Error()
		log.Error(errMsg)
		return common.GeneralError(http.StatusInternalServerError, response.InternalError, errMsg, nil, nil)
	} else if invalidProperties != "" {
		errorMessage := "one or more properties given in the request body are not valid, ensure properties are listed in uppercamelcase "
		log.Error(errorMessage)
		response := common.GeneralError(http.StatusBadRequest, response.PropertyUnknown, errorMessage, []interface{}{invalidProperties}, nil)
		return response
	}

	requestBody, err := json.Marshal(bmcAccReq)
	if err != nil {
		log.Error("while marshalling the update BMC account request: " + err.Error())
		resp = common.GeneralError(http.StatusInternalServerError, response.InternalError, err.Error(), nil, nil)
		return resp
	}
	// splitting managerID to get uuid
	requestData := strings.SplitN(req.ManagerID, ".", 2)
	uuid := requestData[0]

	uri := replaceBMCAccReq(req.URL, req.ManagerID)
	resp = e.deviceCommunication(uri, uuid, requestData[1], http.MethodPatch, requestBody)

	if resp.StatusCode == http.StatusOK {
		data, err := e.getResourceInfoFromDevice(uri, uuid, requestData[1])
		if err != nil {
			errorMessage := "unable to get resource details from device: " + err.Error()
			log.Error(errorMessage)
			errArgs := []interface{}{}
			return common.GeneralError(http.StatusNotFound, response.ResourceNotFound, errorMessage, errArgs, nil)
		}
		// Replace response body to BMC manager
		data = replaceBMCAccResp(data, req.ManagerID)
		resource := convertToRedfishModel(req.URL, data)
		resp.Body = resource
		resp.StatusCode = http.StatusOK
		resp.StatusMessage = response.Success
		//return resp

	}
	return resp

}

// DeleteRemoteAccountService is used to delete the BMC account user
func (e *ExternalInterface) DeleteRemoteAccountService(req *managersproto.ManagerRequest) response.RPC {
	var resp response.RPC
	// splitting managerID to get uuid
	requestData := strings.SplitN(req.ManagerID, ".", 2)
	uuid := requestData[0]
	uri := replaceBMCAccReq(req.URL, req.ManagerID)
	resp = e.deviceCommunication(uri, uuid, requestData[1], http.MethodDelete, nil)
	if resp.StatusCode == http.StatusOK {
		resp.StatusCode = http.StatusNoContent
	}
	return resp
}
