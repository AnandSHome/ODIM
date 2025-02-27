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

package model

import (
	"github.com/ODIM-Project/ODIM/lib-utilities/common"
	"github.com/ODIM-Project/ODIM/lib-utilities/errors"
)

// Chassis redfish structure
type Chassis struct {
	Ocontext                    string            `json:"@odata.context,omitempty"`
	Oid                         string            `json:"@odata.id"`
	Otype                       string            `json:"@odata.type"`
	Oetag                       string            `json:"@odata.etag,omitempty"`
	ID                          string            `json:"Id"`
	Description                 string            `json:"Description,omitempty"`
	Name                        string            `json:"Name"`
	AssetTag                    interface{}       `json:"AssetTag"` // omitempty is not added to make value as null if it's not present
	ChassisType                 string            `json:"ChassisType"`
	DepthMm                     float32           `json:"DepthMm,omitempty"`
	EnvironmentalClass          string            `json:"EnvironmentalClass,omitempty"`
	HeightMm                    float32           `json:"HeightMm,omitempty"`
	IndicatorLED                string            `json:"IndicatorLED,omitempty"`
	Manufacturer                string            `json:"Manufacturer,omitempty"`
	Model                       string            `json:"Model,omitempty"`
	PartNumber                  interface{}       `json:"PartNumber"` // omitempty is not added to make value as null if it's not present
	PowerState                  string            `json:"PowerState,omitempty"`
	SerialNumber                string            `json:"SerialNumber,omitempty"`
	SKU                         string            `json:"SKU,omitempty"`
	UUID                        string            `json:"UUID,omitempty"`
	WeightKg                    float32           `json:"WeightKg,omitempty"`
	WidthMm                     float32           `json:"WidthMm,omitempty"`
	Links                       *Links            `json:"Links,omitempty"`
	Location                    *Link             `json:"Location,omitempty"`
	LogServices                 *LogServices      `json:"LogServices,omitempty"`
	Assembly                    *Assembly         `json:"Assembly,omitempty"`
	NetworkAdapters             *NetworkAdapters  `json:"NetworkAdapters,omitempty"`
	PCIeSlots                   *PCIeSlots        `json:"PCIeSlots,omitempty"`
	PhysicalSecurity            *PhysicalSecurity `json:"PhysicalSecurity,omitempty"`
	Power                       *Link             `json:"Power,omitempty"`
	Sensors                     *Sensors          `json:"Sensors,omitempty"`
	Status                      *Status           `json:"Status,omitempty"`
	Thermal                     *Link             `json:"Thermal,omitempty"`
	Actions                     *OemActions       `json:"Actions,omitempty"`
	Certificates                *Certificates     `json:"Certificates,omitempty"`
	Controls                    *Link             `json:"Controls,omitempty"`
	Drives                      *Link             `json:"Drives,omitempty"`
	EnvironmentMetrics          *Link             `json:"EnvironmentMetrics,omitempty"`
	LocationIndicatorActive     bool              `json:"LocationIndicatorActive,omitempty"`
	MaxPowerWatts               float32           `json:"MaxPowerWatts,omitempty"`
	Measurements                []*Link           `json:"Measurements,omitempty"` // Deprecated in version v1.19.0
	MediaControllers            *Link             `json:"MediaControllers,omitempty"`
	Memory                      *Link             `json:"Memory,omitempty"`
	MemoryDomains               *Link             `json:"MemoryDomains,omitempty"`
	MinPowerWatts               float32           `json:"MinPowerWatts,omitempty"`
	Oem                         *Oem              `json:"Oem,omitempty"`
	PCIeDevices                 *Link             `json:"PCIeDevices,omitempty"`
	PowerSubsystem              *Link             `json:"PowerSubsystem,omitempty"`
	SparePartNumber             string            `json:"SparePartNumber,omitempty"`
	ThermalSubsystem            *Link             `json:"ThermalSubsystem,omitempty"`
	ThermalDirection            string            `json:"ThermalDirection,omitempty"`
	ThermalManagedByParent      bool              `json:"ThermalManagedByParent.omitempty"`
	PoweredByParent             bool              `json:"PoweredByParent,omitempty"`
	Fans                        []*Link           `json:"Fans,omitempty"`
	PowerSupplies               []*Link           `json:"PowerSupplies,omitempty"`
	PowerDistribution           *Link             `json:"PowerDistribution,omitempty"`
	FabricAdapters              *Link             `json:"FabricAdapters,omitempty"`
	ElectricalSourceManagerURIs []string          `json:"ElectricalSourceManagerURIs,omitempty"`
	ElectricalSourceNames       []string          `json:"ElectricalSourceNames,omitempty"`
}

// LogServices get
/*
/redfish/v1/Managers/{ManagerId}/LogServices/{LogServiceId}
/redfish/v1/Systems/{ComputerSystemId}/LogServices/{LogServiceId}
*/
type LogServices struct {
	Oid                 string      `json:"@odata.id"`
	Ocontext            string      `json:"@odata.context,omitempty"`
	Otype               string      `json:"@odata.type"`
	Oetag               string      `json:"@odata.etag,omitempty"`
	ID                  string      `json:"Id"`
	Description         string      `json:"Description,omitempty"`
	Name                string      `json:"Name"`
	DateTime            string      `json:"DateTime,omitempty"`
	DateTimeLocalOffset string      `json:"DateTimeLocalOffset,omitempty"`
	Entries             *Entries    `json:"Entries,omitempty"`
	LogEntryType        string      `json:"LogEntryType,omitempty"`
	MaxNumberOfRecords  int         `json:"MaxNumberOfRecords,omitempty"`
	OverWritePolicy     string      `json:"OverWritePolicy,omitempty"`
	ServiceEnabled      bool        `json:"ServiceEnabled,omitempty"`
	Status              *Status     `json:"Status,omitempty"`
	AutoDSTEnabled      bool        `json:"AutoDSTEnabled,omitempty"`
	Actions             *OemActions `json:"Actions,omitempty"`
	Oem                 *Oem        `json:"Oem,omitempty"`
	SyslogFilters       *SysLog     `json:"SyslogFilters,omitempty"`
}

// SysLog redfish structure
type SysLog struct {
	LogFacilities  []string `json:"LogFacilities,omitempty"`
	LowestSeverity string   `json:"LowestSeverity,omitempty"`
}

//Entries redfish structure
type Entries struct {
	Oid string `json:"@odata.id"`
}

// Assembly redfish structure
type Assembly struct {
	Oid string `json:"@odata.id"`
}

// NetworkAdapters redfish structure
type NetworkAdapters struct {
	Oid string `json:"@odata.id"`
}

// PCIeSlots redfish structure
type PCIeSlots struct {
	Oid string `json:"@odata.id"`
}

// PhysicalSecurity redfish structure
type PhysicalSecurity struct {
	IntrusionSensor       string
	IntrusionSensorNumber int
	IntrusionSensorReArm  string
}

// Sensors redfish structure
type Sensors struct {
	Oid string `json:"@odata.id"`
}

// Status redfish structure
type Status struct {
	Oid          string `json:"@odata.id,omitempty"`
	Ocontext     string `json:"@odata.context,omitempty"`
	Oetag        string `json:"@odata.etag,omitempty"`
	Otype        string `json:"@odata.type,omitempty"`
	Description  string `json:"description,omitempty"`
	ID           string `json:"Id,omitempty"`
	Name         string `json:"Name,omitempty"`
	Health       string `json:"Health,omitempty"`
	HealthRollup string `json:"HealthRollup,omitempty"`
	State        string `json:"State,omitempty"`
	Oem          *Oem   `json:"Oem,omitempty"`
}

// SaveInMemory will create the Chassis in inmemory DB, with key as UUID
// Takes:
//	none as function parameter, but takes c of type *Chassis as a pointer receiver implicitly.
// Returns:
//	err of type error
//
//	On Sucess  - returns nil value
//	On Failure - returns non nil value
func (c *Chassis) SaveInMemory(deviceUUID string) *errors.Error {
	connPool, err := common.GetDBConnection(common.InMemory)
	if err != nil {
		return errors.PackError(err.ErrNo(), "error while trying to connect to DB: ", err.Error())
	}
	if err := connPool.Create("chassis", deviceUUID, c); err != nil {
		return errors.PackError(err.ErrNo(), "error while trying to create new chassis: ", err.Error())
	}
	return nil
}
