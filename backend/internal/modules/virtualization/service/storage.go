package service

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/digitalocean/go-libvirt"
)

type StoragePoolInfo struct {
	Name            string `json:"name"`
	UUID            string `json:"uuid"`
	State           string `json:"state"`
	CapacityBytes   uint64 `json:"capacityBytes"`
	AllocationBytes uint64 `json:"allocationBytes"`
	AvailableBytes  uint64 `json:"availableBytes"`
	Autostart       bool   `json:"autostart"`
	VolumeCount     int    `json:"volumeCount"`
}

type StorageVolumeInfo struct {
	Pool            string `json:"pool"`
	Name            string `json:"name"`
	Path            string `json:"path"`
	Type            string `json:"type"`
	Format          string `json:"format"`
	CapacityBytes   uint64 `json:"capacityBytes"`
	AllocationBytes uint64 `json:"allocationBytes"`
}

type CreateStorageVolumeRequest struct {
	Pool    string `json:"pool" binding:"required"`
	Name    string `json:"name" binding:"required"`
	SizeGiB uint64 `json:"sizeGiB" binding:"required"`
	Format  string `json:"format"`
}

type ResizeStorageVolumeRequest struct {
	SizeGiB uint64 `json:"sizeGiB" binding:"required"`
}

type volumeXML struct {
	Target struct {
		Format struct {
			Type string `xml:"type,attr"`
		} `xml:"format"`
	} `xml:"target"`
}

func (s *Service) ListStoragePools() ([]StoragePoolInfo, error) {
	var result []StoragePoolInfo
	err := s.withClient(func(client *libvirt.Libvirt) error {
		pools, _, err := client.ConnectListAllStoragePools(1, 0)
		if err != nil {
			return err
		}
		result = make([]StoragePoolInfo, 0, len(pools))
		for _, pool := range pools {
			state, capacity, allocation, available, err := client.StoragePoolGetInfo(pool)
			if err != nil {
				return fmt.Errorf("get storage pool %q info: %w", pool.Name, err)
			}
			autostart, err := client.StoragePoolGetAutostart(pool)
			if err != nil {
				return fmt.Errorf("get storage pool %q autostart: %w", pool.Name, err)
			}
			volumes, _, err := client.StoragePoolListAllVolumes(pool, 1, 0)
			if err != nil {
				return fmt.Errorf("list storage pool %q volumes: %w", pool.Name, err)
			}
			result = append(result, StoragePoolInfo{
				Name:            pool.Name,
				UUID:            fmt.Sprintf("%x", pool.UUID),
				State:           storagePoolState(state),
				CapacityBytes:   capacity,
				AllocationBytes: allocation,
				AvailableBytes:  available,
				Autostart:       autostart != 0,
				VolumeCount:     len(volumes),
			})
		}
		return nil
	})
	return result, err
}

func (s *Service) ListStorageVolumes(poolName string) ([]StorageVolumeInfo, error) {
	poolName = strings.TrimSpace(poolName)
	if poolName != "" && !isSafeValue(poolName) {
		return nil, errors.New("invalid storage pool")
	}
	var result []StorageVolumeInfo
	err := s.withClient(func(client *libvirt.Libvirt) error {
		var pools []libvirt.StoragePool
		if poolName != "" {
			pool, err := client.StoragePoolLookupByName(poolName)
			if err != nil {
				return fmt.Errorf("lookup storage pool: %w", err)
			}
			pools = []libvirt.StoragePool{pool}
		} else {
			var err error
			pools, _, err = client.ConnectListAllStoragePools(1, 0)
			if err != nil {
				return err
			}
		}
		result = make([]StorageVolumeInfo, 0)
		for _, pool := range pools {
			volumes, _, err := client.StoragePoolListAllVolumes(pool, 1, 0)
			if err != nil {
				return fmt.Errorf("list storage pool %q volumes: %w", pool.Name, err)
			}
			for _, volume := range volumes {
				info, err := describeStorageVolume(client, pool.Name, volume)
				if err != nil {
					return err
				}
				result = append(result, info)
			}
		}
		return nil
	})
	return result, err
}

func (s *Service) CreateStorageVolume(req CreateStorageVolumeRequest) (StorageVolumeInfo, error) {
	req.Pool = strings.TrimSpace(req.Pool)
	req.Name = strings.TrimSpace(req.Name)
	req.Format = strings.ToLower(strings.TrimSpace(req.Format))
	if !isSafeValue(req.Pool) {
		return StorageVolumeInfo{}, errors.New("invalid storage pool")
	}
	if err := validateVolumeName(req.Name); err != nil {
		return StorageVolumeInfo{}, err
	}
	if req.SizeGiB < 1 || req.SizeGiB > 4096 {
		return StorageVolumeInfo{}, errors.New("sizeGiB must be between 1 and 4096")
	}
	if req.Format == "" {
		req.Format = "qcow2"
	}
	if req.Format != "qcow2" && req.Format != "raw" {
		return StorageVolumeInfo{}, errors.New("format must be qcow2 or raw")
	}

	var result StorageVolumeInfo
	err := s.withClient(func(client *libvirt.Libvirt) error {
		pool, err := client.StoragePoolLookupByName(req.Pool)
		if err != nil {
			return fmt.Errorf("lookup storage pool: %w", err)
		}
		if _, err := client.StorageVolLookupByName(pool, req.Name); err == nil {
			return fmt.Errorf("storage volume %q already exists", req.Name)
		}
		volumeXML := fmt.Sprintf("<volume><name>%s</name><capacity unit='bytes'>%d</capacity><target><format type='%s'/></target></volume>", xmlEscape(req.Name), req.SizeGiB*1024*1024*1024, req.Format)
		volume, err := client.StorageVolCreateXML(pool, volumeXML, 0)
		if err != nil {
			return fmt.Errorf("create storage volume: %w", err)
		}
		result, err = describeStorageVolume(client, req.Pool, volume)
		return err
	})
	return result, err
}

func (s *Service) ResizeStorageVolume(poolName, volumeName string, req ResizeStorageVolumeRequest) (StorageVolumeInfo, error) {
	if !isSafeValue(poolName) {
		return StorageVolumeInfo{}, errors.New("invalid storage pool")
	}
	if err := validateVolumeName(volumeName); err != nil {
		return StorageVolumeInfo{}, err
	}
	if req.SizeGiB < 1 || req.SizeGiB > 4096 {
		return StorageVolumeInfo{}, errors.New("sizeGiB must be between 1 and 4096")
	}
	var result StorageVolumeInfo
	err := s.withClient(func(client *libvirt.Libvirt) error {
		pool, err := client.StoragePoolLookupByName(poolName)
		if err != nil {
			return fmt.Errorf("lookup storage pool: %w", err)
		}
		volume, err := client.StorageVolLookupByName(pool, volumeName)
		if err != nil {
			return fmt.Errorf("lookup storage volume: %w", err)
		}
		_, currentCapacity, _, err := client.StorageVolGetInfo(volume)
		if err != nil {
			return fmt.Errorf("get storage volume info: %w", err)
		}
		newCapacity := req.SizeGiB * 1024 * 1024 * 1024
		if newCapacity < currentCapacity {
			return errors.New("volume shrinking is not supported; choose a larger size")
		}
		if newCapacity > currentCapacity {
			if err := client.StorageVolResize(volume, newCapacity, 0); err != nil {
				return fmt.Errorf("resize storage volume: %w", err)
			}
		}
		result, err = describeStorageVolume(client, poolName, volume)
		return err
	})
	return result, err
}

func (s *Service) DeleteStorageVolume(poolName, volumeName string) error {
	if !isSafeValue(poolName) {
		return errors.New("invalid storage pool")
	}
	if err := validateVolumeName(volumeName); err != nil {
		return err
	}
	return s.withClient(func(client *libvirt.Libvirt) error {
		pool, err := client.StoragePoolLookupByName(poolName)
		if err != nil {
			return fmt.Errorf("lookup storage pool: %w", err)
		}
		volume, err := client.StorageVolLookupByName(pool, volumeName)
		if err != nil {
			return fmt.Errorf("lookup storage volume: %w", err)
		}
		path, err := client.StorageVolGetPath(volume)
		if err != nil {
			return fmt.Errorf("get storage volume path: %w", err)
		}
		domains, _, err := client.ConnectListAllDomains(1, 0)
		if err != nil {
			return fmt.Errorf("list virtual machines: %w", err)
		}
		for _, domain := range domains {
			xmlText, err := client.DomainGetXMLDesc(domain, 0)
			if err != nil {
				return fmt.Errorf("read virtual machine definition: %w", err)
			}
			var parsed domainXML
			if err := xml.Unmarshal([]byte(xmlText), &parsed); err != nil {
				return fmt.Errorf("parse virtual machine definition: %w", err)
			}
			for _, disk := range parsed.Devices.Disks {
				if disk.Source.File == path || disk.Source.Dev == path {
					return fmt.Errorf("storage volume is attached to virtual machine %q", parsed.Name)
				}
			}
		}
		if err := client.StorageVolDelete(volume, 0); err != nil {
			return fmt.Errorf("delete storage volume: %w", err)
		}
		return nil
	})
}

func describeStorageVolume(client *libvirt.Libvirt, poolName string, volume libvirt.StorageVol) (StorageVolumeInfo, error) {
	volumeType, capacity, allocation, err := client.StorageVolGetInfo(volume)
	if err != nil {
		return StorageVolumeInfo{}, fmt.Errorf("get storage volume info: %w", err)
	}
	path, err := client.StorageVolGetPath(volume)
	if err != nil {
		return StorageVolumeInfo{}, fmt.Errorf("get storage volume path: %w", err)
	}
	xmlText, err := client.StorageVolGetXMLDesc(volume, 0)
	if err != nil {
		return StorageVolumeInfo{}, fmt.Errorf("get storage volume definition: %w", err)
	}
	var parsed volumeXML
	if err := xml.Unmarshal([]byte(xmlText), &parsed); err != nil {
		return StorageVolumeInfo{}, fmt.Errorf("parse storage volume definition: %w", err)
	}
	return StorageVolumeInfo{
		Pool:            poolName,
		Name:            volume.Name,
		Path:            path,
		Type:            storageVolumeType(volumeType),
		Format:          parsed.Target.Format.Type,
		CapacityBytes:   capacity,
		AllocationBytes: allocation,
	}, nil
}

func storagePoolState(state uint8) string {
	switch libvirt.StoragePoolState(state) {
	case libvirt.StoragePoolRunning:
		return "running"
	case libvirt.StoragePoolBuilding:
		return "building"
	case libvirt.StoragePoolDegraded:
		return "degraded"
	case libvirt.StoragePoolInaccessible:
		return "inaccessible"
	default:
		return "inactive"
	}
}

func storageVolumeType(volumeType int8) string {
	switch libvirt.StorageVolType(volumeType) {
	case libvirt.StorageVolBlock:
		return "block"
	case libvirt.StorageVolDir:
		return "dir"
	case libvirt.StorageVolNetwork:
		return "network"
	case libvirt.StorageVolNetdir:
		return "netdir"
	case libvirt.StorageVolPloop:
		return "ploop"
	default:
		return "file"
	}
}

func validateVolumeName(name string) error {
	if name == "" || len(name) > 128 || !isSafeValue(name) {
		return errors.New("invalid storage volume name")
	}
	return nil
}
