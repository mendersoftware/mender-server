package model

type DeviceStatistics struct {
	ByStatus DeviceStatisticsByStatus `json:"devices_by_status"`
}

type DeviceStatisticsByStatus struct {
	Accepted DeviceCountPerTier `json:"accepted"`
	Pending  DeviceCountPerTier `json:"pending"`
}

type DeviceCountPerTier struct {
	Standard uint64 `json:"standard"`
	Micro    uint64 `json:"micro"`
	System   uint64 `json:"system"`
}
