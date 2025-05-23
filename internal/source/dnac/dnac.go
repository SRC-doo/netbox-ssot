package dnac

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	dnac "github.com/cisco-en-programmability/dnacenter-go-sdk/v7/sdk"
	"github.com/src-doo/netbox-ssot/internal/netbox/inventory"
	"github.com/src-doo/netbox-ssot/internal/source/common"
	"github.com/src-doo/netbox-ssot/internal/utils"
)

//nolint:revive
type DnacSource struct {
	common.Config

	// Dnac fetched data. Initialized in init functions.
	Sites                           map[string]dnac.ResponseSitesGetSiteResponse                // SiteID -> Site
	Devices                         map[string]dnac.ResponseDevicesGetDeviceListResponse        // DeviceID -> Device
	Interfaces                      map[string]dnac.ResponseDevicesGetAllInterfacesResponse     // InterfaceID -> Interface
	Vlans                           map[int]dnac.ResponseDevicesGetDeviceInterfaceVLANsResponse // VlanID -> Vlan
	WirelessLANInterfaceName2VlanID map[string]int                                              // InterfaceName -> VlanID
	SSID2WirelessProfileDetails     map[string]dnac.ResponseItemWirelessGetWirelessProfileProfileDetailsSSIDDetails
	// SSID2WlanGroupName SSID -> WirelessLANGroup name
	SSID2WlanGroupName map[string]string
	// SSID2SecurityDetails WirelessLANName -> SSIDDetails
	SSID2SecurityDetails map[string]dnac.ResponseItemWirelessGetEnterpriseSSIDSSIDDetails

	// Relations between dnac data. Initialized in init functions.
	Site2Parent           map[string]string          // Site ID -> Parent Site ID
	Site2Devices          map[string]map[string]bool // Site ID - > set of device IDs
	Device2Site           map[string]string          // Device ID -> Site ID
	DeviceID2InterfaceIDs map[string][]string        // DeviceID -> []InterfaceID

	// Netbox related data for easier access. Initialized in sync functions.
	// DeviceID2isMissingPrimaryIP stores devices without primary IP. See ds.syncMissingDevicePrimaryIPs
	DeviceID2isMissingPrimaryIP sync.Map
	// VID2nbVlan: VlanID -> nbVlan
	VID2nbVlan sync.Map
	// SiteID2nbSite: SiteID -> nbSite
	SiteID2nbSite           sync.Map
	DeviceID2nbDevice       sync.Map // DeviceID -> nbDevice
	InterfaceID2nbInterface sync.Map // InterfaceID -> nbInterface
}

func (ds *DnacSource) Init() error {
	dnacURL := fmt.Sprintf(
		"%s://%s:%d",
		ds.Config.SourceConfig.HTTPScheme,
		ds.Config.SourceConfig.Hostname,
		ds.Config.SourceConfig.Port,
	)
	Client, err := dnac.NewClientWithOptions(
		dnacURL,
		ds.SourceConfig.Username,
		ds.SourceConfig.Password,
		"false",
		strconv.FormatBool(ds.SourceConfig.ValidateCert),
		nil,
	)
	if err != nil {
		return fmt.Errorf("creating dnac client: %s", err)
	}
	// Initialize items from vsphere API to local storage
	initFunctions := []func(*dnac.Client) error{
		ds.initSites,
		ds.initMemberships,
		ds.initDevices,
		ds.initInterfaces,
		ds.initWirelessLANs,
	}

	for _, initFunc := range initFunctions {
		startTime := time.Now()
		if err := initFunc(Client); err != nil {
			return fmt.Errorf("dnac initialization failure: %v", err)
		}
		duration := time.Since(startTime)
		ds.Logger.Infof(
			ds.Ctx,
			"Successfully initialized %s in %f seconds",
			utils.ExtractFunctionNameWithTrimPrefix(initFunc, "init"),
			duration.Seconds(),
		)
	}
	return nil
}

func (ds *DnacSource) Sync(nbi *inventory.NetboxInventory) error {
	syncFunctions := []func(*inventory.NetboxInventory) error{
		ds.syncSites,
		ds.syncVlans,
		ds.syncDevices,
		ds.syncDeviceInterfaces,
		ds.syncWirelessLANs,
		ds.syncMissingDevicePrimaryIPs,
	}

	for _, syncFunc := range syncFunctions {
		startTime := time.Now()
		err := syncFunc(nbi)
		if err != nil {
			return err
		}
		duration := time.Since(startTime)
		ds.Logger.Infof(
			ds.Ctx,
			"Successfully synced %s in %f seconds",
			utils.ExtractFunctionNameWithTrimPrefix(syncFunc, "sync"),
			duration.Seconds(),
		)
	}
	return nil
}
