package service

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/digitalocean/go-libvirt"
	"go.uber.org/zap"
)

const defaultURI = "qemu:///system"

var ErrInvalidName = errors.New("invalid virtual machine name")

type Service struct {
	uri    string
	logger *zap.Logger
}

type VirtualMachine struct {
	Name      string   `json:"name"`
	UUID      string   `json:"uuid"`
	State     string   `json:"state"`
	MemoryMiB uint64   `json:"memoryMiB"`
	VCPUs     uint32   `json:"vcpus"`
	Autostart bool     `json:"autostart"`
	DiskPaths []string `json:"diskPaths"`
	Network   string   `json:"network,omitempty"`
	XML       string   `json:"xml,omitempty"`
}

type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	MemoryMiB   uint64 `json:"memoryMiB"`
	VCPUs       uint32 `json:"vcpus"`
	DiskPath    string `json:"diskPath"`
	DiskSizeGiB uint64 `json:"diskSizeGiB"`
	DiskFormat  string `json:"diskFormat"`
	Pool        string `json:"pool"`
	Network     string `json:"network"`
	Autostart   bool   `json:"autostart"`
	Start       bool   `json:"start"`
}

type UpdateRequest struct {
	MemoryMiB *uint64 `json:"memoryMiB"`
	VCPUs     *uint32 `json:"vcpus"`
	Autostart *bool   `json:"autostart"`
}

type domainXML struct {
	Name          string         `xml:"name"`
	Memory        memoryElement  `xml:"memory"`
	CurrentMemory memoryElement  `xml:"currentMemory"`
	VCPU          vcpuElement    `xml:"vcpu"`
	Devices       devicesElement `xml:"devices"`
}

type memoryElement struct {
	Unit  string `xml:"unit,attr"`
	Value uint64 `xml:",chardata"`
}

type vcpuElement struct {
	Value uint32 `xml:",chardata"`
}

type devicesElement struct {
	Disks      []diskElement      `xml:"disk"`
	Interfaces []interfaceElement `xml:"interface"`
}

type diskElement struct {
	Device string        `xml:"device,attr"`
	Source sourceElement `xml:"source"`
}

type interfaceElement struct {
	Source networkSourceElement `xml:"source"`
}

type sourceElement struct {
	File string `xml:"file,attr"`
	Dev  string `xml:"dev,attr"`
}

type networkSourceElement struct {
	Network string `xml:"network,attr"`
}

func NewFromEnv(logger *zap.Logger) *Service {
	uri := strings.TrimSpace(os.Getenv("LIBVIRT_URI"))
	if uri == "" {
		uri = defaultURI
	}
	return &Service{uri: uri, logger: logger}
}

func (s *Service) withClient(fn func(*libvirt.Libvirt) error) error {
	parsed, err := url.Parse(s.uri)
	if err != nil {
		return fmt.Errorf("parse libvirt URI: %w", err)
	}
	client, err := libvirt.ConnectToURI(parsed)
	if err != nil {
		return fmt.Errorf("connect to libvirt: %w", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil && s.logger != nil {
			s.logger.Warn("disconnect libvirt failed", zap.Error(err))
		}
	}()
	return fn(client)
}

func (s *Service) List() ([]VirtualMachine, error) {
	var result []VirtualMachine
	err := s.withClient(func(client *libvirt.Libvirt) error {
		domains, _, err := client.ConnectListAllDomains(1, 0)
		if err != nil {
			return err
		}
		result = make([]VirtualMachine, 0, len(domains))
		for _, domain := range domains {
			vm, err := s.describe(client, domain, false)
			if err != nil {
				return err
			}
			result = append(result, vm)
		}
		return nil
	})
	return result, err
}

func (s *Service) Get(name string, includeXML bool) (VirtualMachine, error) {
	if err := validateName(name); err != nil {
		return VirtualMachine{}, err
	}
	var result VirtualMachine
	err := s.withClient(func(client *libvirt.Libvirt) error {
		domain, err := client.DomainLookupByName(name)
		if err != nil {
			return fmt.Errorf("lookup virtual machine: %w", err)
		}
		result, err = s.describe(client, domain, includeXML)
		return err
	})
	return result, err
}

func (s *Service) Create(req CreateRequest) (VirtualMachine, error) {
	if err := validateCreateRequest(&req); err != nil {
		return VirtualMachine{}, err
	}
	var result VirtualMachine
	err := s.withClient(func(client *libvirt.Libvirt) error {
		if _, err := client.DomainLookupByName(req.Name); err == nil {
			return fmt.Errorf("virtual machine %q already exists", req.Name)
		}

		diskPath := strings.TrimSpace(req.DiskPath)
		createdVolume := libvirt.StorageVol{}
		volumeCreated := false
		if diskPath == "" && req.DiskSizeGiB > 0 {
			var err error
			diskPath, createdVolume, err = createVolume(client, req.Pool, req.Name+".qcow2", req.DiskSizeGiB)
			if err != nil {
				return err
			}
			volumeCreated = true
		}
		xmlText := buildDomainXML(req, diskPath)
		domain, err := client.DomainDefineXML(xmlText)
		if err != nil {
			if volumeCreated {
				_ = client.StorageVolDelete(createdVolume, 0)
			}
			return fmt.Errorf("define virtual machine: %w", err)
		}
		if req.Autostart {
			if err := client.DomainSetAutostart(domain, 1); err != nil {
				return fmt.Errorf("set autostart: %w", err)
			}
		}
		if req.Start {
			if err := client.DomainCreate(domain); err != nil {
				return fmt.Errorf("start virtual machine: %w", err)
			}
		}
		result, err = s.describe(client, domain, false)
		return err
	})
	return result, err
}

func (s *Service) Update(name string, req UpdateRequest) (VirtualMachine, error) {
	if err := validateName(name); err != nil {
		return VirtualMachine{}, err
	}
	if req.MemoryMiB != nil && (*req.MemoryMiB < 128 || *req.MemoryMiB > 1048576) {
		return VirtualMachine{}, errors.New("memoryMiB must be between 128 and 1048576")
	}
	if req.VCPUs != nil && (*req.VCPUs < 1 || *req.VCPUs > 256) {
		return VirtualMachine{}, errors.New("vcpus must be between 1 and 256")
	}
	var result VirtualMachine
	err := s.withClient(func(client *libvirt.Libvirt) error {
		domain, err := client.DomainLookupByName(name)
		if err != nil {
			return fmt.Errorf("lookup virtual machine: %w", err)
		}
		if req.MemoryMiB == nil && req.VCPUs == nil && req.Autostart == nil {
			result, err = s.describe(client, domain, false)
			return err
		}
		xmlText, err := client.DomainGetXMLDesc(domain, 0)
		if err != nil {
			return fmt.Errorf("read virtual machine definition: %w", err)
		}
		if req.MemoryMiB != nil {
			xmlText, err = rewriteElement(xmlText, "memory", strconv.FormatUint(*req.MemoryMiB*1024, 10))
			if err == nil {
				xmlText, err = rewriteElement(xmlText, "currentMemory", strconv.FormatUint(*req.MemoryMiB*1024, 10))
			}
		}
		if err == nil && req.VCPUs != nil {
			xmlText, err = rewriteElement(xmlText, "vcpu", strconv.FormatUint(uint64(*req.VCPUs), 10))
		}
		if err != nil {
			return fmt.Errorf("update virtual machine definition: %w", err)
		}
		updated, err := client.DomainDefineXML(xmlText)
		if err != nil {
			return fmt.Errorf("save virtual machine definition: %w", err)
		}
		active, err := client.DomainIsActive(domain)
		if err != nil {
			return fmt.Errorf("read virtual machine state: %w", err)
		}
		if active != 0 && req.MemoryMiB != nil {
			if err := client.DomainSetMemoryFlags(updated, *req.MemoryMiB*1024, uint32(libvirt.DomainMemLive)); err != nil {
				return fmt.Errorf("apply live memory: %w", err)
			}
		}
		if active != 0 && req.VCPUs != nil {
			if err := client.DomainSetVcpusFlags(updated, *req.VCPUs, uint32(libvirt.DomainVCPULive)); err != nil {
				return fmt.Errorf("apply live vcpus: %w", err)
			}
		}
		if req.Autostart != nil {
			value := int32(0)
			if *req.Autostart {
				value = 1
			}
			if err := client.DomainSetAutostart(updated, value); err != nil {
				return fmt.Errorf("update autostart: %w", err)
			}
		}
		result, err = s.describe(client, updated, false)
		return err
	})
	return result, err
}

func (s *Service) Start(name string) error {
	return s.withDomain(name, func(client *libvirt.Libvirt, domain libvirt.Domain) error {
		return client.DomainCreate(domain)
	})
}

func (s *Service) Shutdown(name string) error {
	return s.withDomain(name, func(client *libvirt.Libvirt, domain libvirt.Domain) error {
		return client.DomainShutdown(domain)
	})
}

func (s *Service) Reboot(name string) error {
	return s.withDomain(name, func(client *libvirt.Libvirt, domain libvirt.Domain) error {
		return client.DomainReboot(domain, libvirt.DomainRebootDefault)
	})
}

func (s *Service) ForceStop(name string) error {
	return s.withDomain(name, func(client *libvirt.Libvirt, domain libvirt.Domain) error {
		return client.DomainDestroy(domain)
	})
}

func (s *Service) Delete(name string) error {
	return s.withDomain(name, func(client *libvirt.Libvirt, domain libvirt.Domain) error {
		active, err := client.DomainIsActive(domain)
		if err != nil {
			return err
		}
		if active != 0 {
			return errors.New("virtual machine is running; shut it down before deleting")
		}
		return client.DomainUndefine(domain)
	})
}

func (s *Service) withDomain(name string, fn func(*libvirt.Libvirt, libvirt.Domain) error) error {
	if err := validateName(name); err != nil {
		return err
	}
	return s.withClient(func(client *libvirt.Libvirt) error {
		domain, err := client.DomainLookupByName(name)
		if err != nil {
			return fmt.Errorf("lookup virtual machine: %w", err)
		}
		return fn(client, domain)
	})
}

func (s *Service) describe(client *libvirt.Libvirt, domain libvirt.Domain, includeXML bool) (VirtualMachine, error) {
	state, _, _, _, _, err := client.DomainGetInfo(domain)
	if err != nil {
		return VirtualMachine{}, err
	}
	xmlText, err := client.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return VirtualMachine{}, err
	}
	var parsed domainXML
	if err := xml.Unmarshal([]byte(xmlText), &parsed); err != nil {
		return VirtualMachine{}, err
	}
	autostart, err := client.DomainGetAutostart(domain)
	if err != nil {
		return VirtualMachine{}, err
	}
	memory := parsed.CurrentMemory.Value
	if memory == 0 {
		memory = parsed.Memory.Value
	}
	memoryMiB := memoryToMiB(memory, parsed.CurrentMemory.Unit)
	if memoryMiB == 0 {
		memoryMiB = memoryToMiB(parsed.Memory.Value, parsed.Memory.Unit)
	}
	disks := make([]string, 0, len(parsed.Devices.Disks))
	for _, disk := range parsed.Devices.Disks {
		if disk.Source.File != "" {
			disks = append(disks, disk.Source.File)
		}
	}
	network := ""
	for _, iface := range parsed.Devices.Interfaces {
		if iface.Source.Network != "" {
			network = iface.Source.Network
			break
		}
	}
	vm := VirtualMachine{
		Name:      parsed.Name,
		UUID:      fmt.Sprintf("%x", domain.UUID),
		State:     domainState(state),
		MemoryMiB: memoryMiB,
		VCPUs:     parsed.VCPU.Value,
		Autostart: autostart != 0,
		DiskPaths: disks,
		Network:   network,
	}
	if includeXML {
		vm.XML = xmlText
	}
	return vm, nil
}

func domainState(state uint8) string {
	switch libvirt.DomainState(state) {
	case libvirt.DomainRunning:
		return "running"
	case libvirt.DomainBlocked:
		return "blocked"
	case libvirt.DomainPaused:
		return "paused"
	case libvirt.DomainShutdown:
		return "shutdown"
	case libvirt.DomainShutoff:
		return "shutoff"
	case libvirt.DomainCrashed:
		return "crashed"
	case libvirt.DomainPmsuspended:
		return "pmsuspended"
	default:
		return "unknown"
	}
}

func validateCreateRequest(req *CreateRequest) error {
	if err := validateName(req.Name); err != nil {
		return err
	}
	if req.MemoryMiB == 0 {
		req.MemoryMiB = 2048
	}
	if req.MemoryMiB < 128 || req.MemoryMiB > 1048576 {
		return errors.New("memoryMiB must be between 128 and 1048576")
	}
	if req.VCPUs == 0 {
		req.VCPUs = 2
	}
	if req.VCPUs > 256 {
		return errors.New("vcpus must be between 1 and 256")
	}
	if req.DiskPath != "" {
		if !filepath.IsAbs(req.DiskPath) {
			return errors.New("diskPath must be an absolute path")
		}
		allowed := strings.TrimSpace(os.Getenv("LIBVIRT_ALLOWED_DISK_PREFIXES"))
		if allowed == "" {
			allowed = "/var/lib/libvirt/images,/home/ytq"
		}
		ok := false
		for _, prefix := range strings.Split(allowed, ",") {
			prefix = filepath.Clean(strings.TrimSpace(prefix))
			if prefix != "" && (req.DiskPath == prefix || strings.HasPrefix(req.DiskPath, prefix+string(os.PathSeparator))) {
				ok = true
				break
			}
		}
		if !ok {
			return errors.New("diskPath is outside the configured libvirt image directories")
		}
	}
	if req.DiskSizeGiB > 1024 {
		return errors.New("diskSizeGiB must be at most 1024")
	}
	if req.DiskPath != "" && req.DiskSizeGiB > 0 {
		return errors.New("diskPath and diskSizeGiB cannot be used together")
	}
	if req.DiskFormat == "" {
		req.DiskFormat = "qcow2"
	}
	if req.DiskFormat != "qcow2" && req.DiskFormat != "raw" {
		return errors.New("diskFormat must be qcow2 or raw")
	}
	if req.Network == "" {
		req.Network = "default"
	}
	if !isSafeValue(req.Network) {
		return errors.New("invalid network")
	}
	if req.Pool == "" {
		req.Pool = "default"
	}
	if !isSafeValue(req.Pool) {
		return errors.New("invalid storage pool")
	}
	return nil
}

func validateName(name string) error {
	if name == "" || len(name) > 128 || !isSafeValue(name) {
		return ErrInvalidName
	}
	return nil
}

func isSafeValue(value string) bool {
	for _, r := range value {
		if !(r == '-' || r == '_' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return value != ""
}

func buildDomainXML(req CreateRequest, diskPath string) string {
	var b strings.Builder
	b.WriteString("<domain type='kvm'>")
	fmt.Fprintf(&b, "<name>%s</name>", xmlEscape(req.Name))
	fmt.Fprintf(&b, "<memory unit='KiB'>%d</memory><currentMemory unit='KiB'>%d</currentMemory>", req.MemoryMiB*1024, req.MemoryMiB*1024)
	fmt.Fprintf(&b, "<vcpu placement='static'>%d</vcpu>", req.VCPUs)
	b.WriteString("<os><type arch='x86_64' machine='q35'>hvm</type><boot dev='hd'/></os><features><acpi/><apic/></features>")
	b.WriteString("<clock offset='utc'/><on_poweroff>destroy</on_poweroff><on_reboot>restart</on_reboot><on_crash>destroy</on_crash>")
	b.WriteString("<devices>")
	if diskPath != "" {
		fmt.Fprintf(&b, "<disk type='file' device='disk'><driver name='qemu' type='%s'/><source file='%s'/><target dev='vda' bus='virtio'/></disk>", xmlEscape(req.DiskFormat), xmlEscape(diskPath))
	}
	fmt.Fprintf(&b, "<interface type='network'><source network='%s'/><model type='virtio'/></interface>", xmlEscape(req.Network))
	b.WriteString("<console type='pty'><target type='serial' port='0'/></console></devices></domain>")
	return b.String()
}

func createVolume(client *libvirt.Libvirt, poolName, volumeName string, sizeGiB uint64) (string, libvirt.StorageVol, error) {
	pool, err := client.StoragePoolLookupByName(poolName)
	if err != nil {
		return "", libvirt.StorageVol{}, fmt.Errorf("lookup storage pool: %w", err)
	}
	volumeXML := fmt.Sprintf("<volume><name>%s</name><capacity unit='bytes'>%d</capacity><target><format type='qcow2'/></target></volume>", xmlEscape(volumeName), sizeGiB*1024*1024*1024)
	volume, err := client.StorageVolCreateXML(pool, volumeXML, 0)
	if err != nil {
		return "", libvirt.StorageVol{}, fmt.Errorf("create disk volume: %w", err)
	}
	path, err := client.StorageVolGetPath(volume)
	if err != nil {
		_ = client.StorageVolDelete(volume, 0)
		return "", libvirt.StorageVol{}, fmt.Errorf("get disk volume path: %w", err)
	}
	return path, volume, nil
}

func rewriteElement(input, element, value string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(input))
	var out bytes.Buffer
	encoder := xml.NewEncoder(&out)
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		if start, ok := token.(xml.StartElement); ok && start.Name.Local == element {
			if err := encoder.EncodeToken(start); err != nil {
				return "", err
			}
			next, err := decoder.Token()
			if err != nil {
				return "", err
			}
			if _, ok := next.(xml.CharData); ok {
				if err := encoder.EncodeToken(xml.CharData([]byte(value))); err != nil {
					return "", err
				}
				continue
			}
			if err := encoder.EncodeToken(next); err != nil {
				return "", err
			}
			continue
		}
		if err := encoder.EncodeToken(token); err != nil {
			return "", err
		}
	}
	if err := encoder.Flush(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func memoryToMiB(value uint64, unit string) uint64 {
	switch strings.ToLower(unit) {
	case "b", "bytes":
		return value / (1024 * 1024)
	case "mib", "mb", "":
		return value
	case "gib", "gb":
		return value * 1024
	default:
		return value / 1024
	}
}

func xmlEscape(value string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}
