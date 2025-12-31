package metrics

type ServerStatus int

const (
	SERVER_STATUS_OFFLINE ServerStatus = iota
	SERVER_STATUS_ONLINE
)

var serverStatusName = map[ServerStatus]string{
	SERVER_STATUS_OFFLINE: "offline",
	SERVER_STATUS_ONLINE:  "online",
}

func (ss ServerStatus) String() string {
	return serverStatusName[ss]
}

type RescueSystemStatus int

const (
	RESCUE_SYSTEM_INACTIVE RescueSystemStatus = iota
	RESCUE_SYSTEM_ACTIVE
)

var rescueSystemStatusName = map[RescueSystemStatus]string{
	RESCUE_SYSTEM_INACTIVE: "inactive",
	RESCUE_SYSTEM_ACTIVE:   "active",
}

func (rss RescueSystemStatus) String() string {
	return rescueSystemStatusName[rss]
}

type RebootRecommendationStatus int

const (
	REBOOT_NOT_RECOMMENDED RebootRecommendationStatus = iota
	REBOOT_RECOMMENDED
)

var rebootRecommendationStatusName = map[RebootRecommendationStatus]string{
	REBOOT_NOT_RECOMMENDED: "not_recommended",
	REBOOT_RECOMMENDED:     "recommended",
}

func (rrs RebootRecommendationStatus) String() string {
	return rebootRecommendationStatusName[rrs]
}

type DiskOptimizationStatus int

const (
	DISK_OPTIMIZATION_NO DiskOptimizationStatus = iota
	DISK_OPTIMIZATION_YES
)

var diskOptimizationStatusName = map[DiskOptimizationStatus]string{
	DISK_OPTIMIZATION_NO:  "no",
	DISK_OPTIMIZATION_YES: "yes",
}

func (dos DiskOptimizationStatus) String() string {
	return diskOptimizationStatusName[dos]
}

type InterfaceThrottledStatus int

const (
	INTERFACE_NOT_THROTTLED InterfaceThrottledStatus = iota
	INTERFACE_THROTTLED
)

var interfaceThrottledStatusName = map[InterfaceThrottledStatus]string{
	INTERFACE_NOT_THROTTLED: "not_throttled",
	INTERFACE_THROTTLED:     "throttled",
}

func (its InterfaceThrottledStatus) String() string {
	return interfaceThrottledStatusName[its]
}
