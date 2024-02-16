package constants

import "github.com/bl4ko/netbox-ssot/internal/netbox/objects"

type SourceType string

const (
	Ovirt  SourceType = "ovirt"
	Vmware SourceType = "vmware"
	Dnac   SourceType = "dnac"
)

// Default mappings of sources to colors (for tags).
var DefaultSourceToTagColorMap = map[SourceType]string{
	Ovirt:  objects.COLOR_DARK_RED,
	Vmware: objects.COLOR_LIGHT_GREEN,
	Dnac:   objects.COLOR_LIGHT_BLUE,
}

// Object for mapping source type to tag color.
var SourceTypeToTagColorMap = map[SourceType]string{
	Ovirt:  objects.COLOR_RED,
	Vmware: objects.COLOR_GREEN,
	Dnac:   objects.COLOR_BLUE,
}

const (
	DefaultTimeout = 10
)

// Magic numbers for dealing with bytes.
const (
	B   = 1
	KB  = 1000 * B
	MB  = 1000 * KB
	GB  = 1000 * MB
	TB  = 1000 * GB
	KiB = 1024 * B
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
)

// Magic numbers for dealing with IP addresses.
const (
	IPv4 = 4
	IPv6 = 6
)

const (
	HTTPSDefaultPort = 443
)
