package proxy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	EngineSingBox = "sing-box"
	EngineXray    = "xray"
	ScopeSystem   = "system"
	ScopeUser     = "user"
)

var unitPattern = regexp.MustCompile("^[A-Za-z0-9_.@:-]+$")

type Config struct {
	DB             *gorm.DB
	BackupDir      string
	ConfigRoots    []string
	CommandTimeout time.Duration
}

type Service struct {
	db             *gorm.DB
	backupDir      string
	configRoots    []string
	commandTimeout time.Duration
	mu             sync.Mutex
}

type Status struct {
	State     string    `json:"state"`
	SubState  string    `json:"subState"`
	MainPID   int       `json:"mainPid"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Metrics struct {
	PID         int       `json:"pid"`
	ReadBytes   uint64    `json:"readBytes"`
	WriteBytes  uint64    `json:"writeBytes"`
	ReadCalls   uint64    `json:"readCalls"`
	WriteCalls  uint64    `json:"writeCalls"`
	Connections int       `json:"connections"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// InstanceView is the public proxy instance shape. The shared BaseModel uses
// an uppercase ID for legacy APIs, while this module exposes the lower camel
// case contract consumed by the proxy manager UI.
type InstanceView struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Engine      string `json:"engine"`
	Scope       string `json:"scope"`
	Unit        string `json:"unit"`
	BinaryPath  string `json:"binaryPath"`
	ConfigPath  string `json:"configPath"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
	Status      Status `json:"status"`
}

func instanceView(instance Instance, status Status) InstanceView {
	return InstanceView{
		ID: instance.ID, Name: instance.Name, Engine: instance.Engine,
		Scope: instance.Scope, Unit: instance.Unit, BinaryPath: instance.BinaryPath,
		ConfigPath: instance.ConfigPath, Enabled: instance.Enabled,
		Description: instance.Description, Status: status,
	}
}

type ConfigSnapshot struct {
	Content  string    `json:"content"`
	Digest   string    `json:"digest"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

type SaveResult struct {
	Digest    string `json:"digest"`
	Backup    string `json:"backup,omitempty"`
	Validated bool   `json:"validated"`
}

func NewService(cfg Config) (*Service, error) {
	if cfg.DB == nil {
		return nil, errors.New("proxy manager requires a database")
	}
	if cfg.BackupDir == "" {
		cfg.BackupDir = "/var/lib/base-frame/proxy-backups"
	}
	if cfg.CommandTimeout <= 0 {
		cfg.CommandTimeout = 15 * time.Second
	}
	if len(cfg.ConfigRoots) == 0 {
		cfg.ConfigRoots = []string{"/etc/sing-box", "/etc/xray", "/home/claude/agsbx", "/var/lib/base-frame/proxy"}
	}
	if err := os.MkdirAll(cfg.BackupDir, 0700); err != nil {
		return nil, fmt.Errorf("create proxy backup directory: %w", err)
	}
	return &Service{db: cfg.DB, backupDir: cfg.BackupDir, configRoots: cfg.ConfigRoots, commandTimeout: cfg.CommandTimeout}, nil
}

func (s *Service) Seed(ctx context.Context) error {
	defaults := []Instance{
		{Name: "sing-box", Engine: EngineSingBox, Scope: ScopeSystem, Unit: "sing-box.service", BinaryPath: "/usr/bin/sing-box", ConfigPath: "/etc/sing-box/config.json", Enabled: true, Description: "sing-box system service"},
		{Name: "Xray", Engine: EngineXray, Scope: ScopeUser, Unit: "xray.service", BinaryPath: "/home/claude/agsbx/xray", ConfigPath: "/home/claude/agsbx/xr.json", Enabled: true, Description: "Xray core service"},
	}
	for _, item := range defaults {
		var existing Instance
		err := s.db.WithContext(ctx).Where("name = ?", item.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.db.WithContext(ctx).Create(&item).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context) ([]InstanceView, error) {
	var instances []Instance
	if err := s.db.WithContext(ctx).Order("id asc").Find(&instances).Error; err != nil {
		return nil, err
	}
	views := make([]InstanceView, 0, len(instances))
	for _, instance := range instances {
		views = append(views, instanceView(instance, s.Status(ctx, instance)))
	}
	return views, nil
}

func (s *Service) Get(ctx context.Context, id uint) (Instance, error) {
	var instance Instance
	if err := s.db.WithContext(ctx).First(&instance, id).Error; err != nil {
		return instance, err
	}
	return instance, nil
}

func (s *Service) Create(ctx context.Context, instance Instance) (Instance, error) {
	if err := validateInstance(instance); err != nil {
		return instance, err
	}
	if err := s.db.WithContext(ctx).Create(&instance).Error; err != nil {
		return instance, err
	}
	return instance, nil
}

func (s *Service) Update(ctx context.Context, instance Instance) (Instance, error) {
	if err := validateInstance(instance); err != nil {
		return instance, err
	}
	var current Instance
	if err := s.db.WithContext(ctx).First(&current, instance.ID).Error; err != nil {
		return instance, err
	}
	if err := s.db.WithContext(ctx).Model(&current).Updates(map[string]any{
		"name": instance.Name, "engine": instance.Engine, "scope": instance.Scope,
		"unit": instance.Unit, "binary_path": instance.BinaryPath, "config_path": instance.ConfigPath,
		"enabled": instance.Enabled, "description": instance.Description,
	}).Error; err != nil {
		return instance, err
	}
	return s.Get(ctx, instance.ID)
}

func (s *Service) Status(ctx context.Context, instance Instance) Status {
	status := Status{State: "unknown", UpdatedAt: time.Now()}
	if err := validateInstance(instance); err != nil {
		status.State, status.Error = "invalid", err.Error()
		return status
	}
	output, err := s.systemctl(ctx, instance.Scope, "show", instance.Unit, "--no-pager", "--property=ActiveState,SubState,MainPID")
	if err != nil {
		status.State, status.Error = "unavailable", err.Error()
		return status
	}
	values := parseProperties(output)
	status.State = values["ActiveState"]
	if status.State == "" {
		status.State = "unknown"
	}
	status.SubState = values["SubState"]
	status.MainPID, _ = strconv.Atoi(values["MainPID"])
	return status
}

func (s *Service) Action(ctx context.Context, id uint, action string) (Status, error) {
	instance, err := s.Get(ctx, id)
	if err != nil {
		return Status{}, err
	}
	switch action {
	case "start", "stop", "restart", "reload", "enable", "disable":
	default:
		return Status{}, fmt.Errorf("unsupported action %q", action)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.systemctl(ctx, instance.Scope, action, instance.Unit); err != nil {
		return Status{}, err
	}
	return s.Status(ctx, instance), nil
}

func (s *Service) Validate(ctx context.Context, id uint) (string, error) {
	instance, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return s.validatePath(ctx, instance, instance.ConfigPath)
}

func (s *Service) ReadConfig(ctx context.Context, id uint) (ConfigSnapshot, error) {
	instance, err := s.Get(ctx, id)
	if err != nil {
		return ConfigSnapshot{}, err
	}
	if err := s.validateConfigPath(instance.ConfigPath); err != nil {
		return ConfigSnapshot{}, err
	}
	info, err := os.Stat(instance.ConfigPath)
	if err != nil {
		return ConfigSnapshot{}, err
	}
	data, err := os.ReadFile(instance.ConfigPath)
	if err != nil {
		return ConfigSnapshot{}, err
	}
	if len(data) > 4<<20 {
		return ConfigSnapshot{}, errors.New("configuration exceeds 4 MiB")
	}
	return ConfigSnapshot{Content: string(data), Digest: digest(data), Size: info.Size(), Modified: info.ModTime()}, nil
}

func (s *Service) SaveConfig(ctx context.Context, id uint, content string) (SaveResult, error) {
	instance, err := s.Get(ctx, id)
	if err != nil {
		return SaveResult{}, err
	}
	if len(content) == 0 || len(content) > 4<<20 {
		return SaveResult{}, errors.New("configuration must be between 1 byte and 4 MiB")
	}
	if err := s.validateConfigPath(instance.ConfigPath); err != nil {
		return SaveResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tmp, err := os.CreateTemp(filepath.Dir(instance.ConfigPath), ".base-frame-proxy-validate-*")
	if err != nil {
		return SaveResult{}, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return SaveResult{}, err
	}
	if err := tmp.Close(); err != nil {
		return SaveResult{}, err
	}
	if _, err := s.validateCommand(ctx, instance, tmpPath); err != nil {
		return SaveResult{}, err
	}
	backupName, err := s.backupCurrent(instance)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return SaveResult{}, err
	}
	mode := os.FileMode(0600)
	if info, statErr := os.Stat(instance.ConfigPath); statErr == nil {
		mode = info.Mode().Perm()
	}
	writePath := instance.ConfigPath + ".next"
	if err := os.WriteFile(writePath, []byte(content), mode); err != nil {
		return SaveResult{}, err
	}
	if err := os.Rename(writePath, instance.ConfigPath); err != nil {
		os.Remove(writePath)
		return SaveResult{}, err
	}
	return SaveResult{Digest: digest([]byte(content)), Backup: backupName, Validated: true}, nil
}

func (s *Service) Rollback(ctx context.Context, id uint) (SaveResult, error) {
	instance, err := s.Get(ctx, id)
	if err != nil {
		return SaveResult{}, err
	}
	if err := s.validateConfigPath(instance.ConfigPath); err != nil {
		return SaveResult{}, err
	}
	matches, err := filepath.Glob(filepath.Join(s.backupDir, fmt.Sprintf("%d-*.bak", id)))
	if err != nil || len(matches) == 0 {
		return SaveResult{}, errors.New("no configuration backup exists")
	}
	latest := matches[len(matches)-1]
	data, err := os.ReadFile(latest)
	if err != nil {
		return SaveResult{}, err
	}
	if _, err := s.validateContent(ctx, instance, data); err != nil {
		return SaveResult{}, err
	}
	currentBackup, _ := s.backupCurrent(instance)
	if err := os.WriteFile(instance.ConfigPath+".next", data, 0600); err != nil {
		return SaveResult{}, err
	}
	if err := os.Rename(instance.ConfigPath+".next", instance.ConfigPath); err != nil {
		os.Remove(instance.ConfigPath + ".next")
		return SaveResult{}, err
	}
	return SaveResult{Digest: digest(data), Backup: currentBackup, Validated: true}, nil
}

func (s *Service) Metrics(ctx context.Context, id uint) (Metrics, error) {
	instance, err := s.Get(ctx, id)
	if err != nil {
		return Metrics{}, err
	}
	status := s.Status(ctx, instance)
	metrics := Metrics{PID: status.MainPID, UpdatedAt: time.Now()}
	if status.MainPID <= 0 {
		return metrics, nil
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/io", status.MainPID))
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				continue
			}
			value, _ := strconv.ParseUint(parts[1], 10, 64)
			switch parts[0] {
			case "read_bytes:":
				metrics.ReadBytes = value
			case "write_bytes:":
				metrics.WriteBytes = value
			case "syscr:":
				metrics.ReadCalls = value
			case "syscw:":
				metrics.WriteCalls = value
			}
		}
	}
	if output, connErr := exec.CommandContext(ctx, "ss", "-Hntup").Output(); connErr == nil {
		needle := fmt.Sprintf("pid=%d,", status.MainPID)
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, needle) {
				metrics.Connections++
			}
		}
	}
	return metrics, nil
}

func (s *Service) validatePath(ctx context.Context, instance Instance, path string) (string, error) {
	if err := s.validateConfigPath(path); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return s.validateCommand(ctx, instance, path)
}

func (s *Service) validateContent(ctx context.Context, instance Instance, data []byte) (string, error) {
	if len(data) == 0 || len(data) > 4<<20 {
		return "", errors.New("configuration size is invalid")
	}
	tmp, err := os.CreateTemp("", "base-frame-proxy-check-*")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	defer os.Remove(path)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	return s.validateCommand(ctx, instance, path)
}

func (s *Service) validateCommand(ctx context.Context, instance Instance, path string) (string, error) {
	var args []string
	switch instance.Engine {
	case EngineSingBox:
		args = []string{"check", "-c", path}
	case EngineXray:
		args = []string{"run", "-test", "-c", path}
	default:
		return "", errors.New("unsupported engine")
	}
	commandCtx, cancel := context.WithTimeout(ctx, s.commandTimeout)
	defer cancel()
	out, err := exec.CommandContext(commandCtx, instance.BinaryPath, args...).CombinedOutput()
	message := strings.TrimSpace(string(out))
	if err != nil {
		if message == "" {
			message = err.Error()
		}
		return "", errors.New(message)
	}
	return message, nil
}

func (s *Service) validateConfigPath(path string) error {
	if !filepath.IsAbs(path) || strings.Contains(filepath.Clean(path), "..") {
		return errors.New("configuration path must be absolute and cannot contain parent traversal")
	}
	clean := filepath.Clean(path)
	for _, root := range s.configRoots {
		root = filepath.Clean(root)
		if clean == root || strings.HasPrefix(clean, root+string(os.PathSeparator)) {
			return nil
		}
	}
	return fmt.Errorf("configuration path is outside allowed roots")
}

func (s *Service) systemctl(ctx context.Context, scope string, args ...string) (string, error) {
	if scope != ScopeSystem && scope != ScopeUser {
		return "", errors.New("unsupported service scope")
	}
	if len(args) == 0 {
		return "", errors.New("systemd action is required")
	}
	unit := ""
	if len(args) > 1 {
		unit = args[1]
	} else {
		unit = args[0]
	}
	if !unitPattern.MatchString(unit) {
		return "", errors.New("invalid systemd unit")
	}
	commandArgs := make([]string, 0, len(args)+2)
	if scope == ScopeUser {
		// Xray currently runs in claude's lingering user manager.
		commandArgs = append(commandArgs, "--user", "-M", "claude@.host")
	}
	commandArgs = append(commandArgs, args...)
	commandCtx, cancel := context.WithTimeout(ctx, s.commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(commandCtx, "/usr/bin/systemctl", commandArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("systemctl %s: %s", args[0], strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *Service) backupCurrent(instance Instance) (string, error) {
	data, err := os.ReadFile(instance.ConfigPath)
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d-%d.bak", instance.ID, time.Now().UnixNano())
	path := filepath.Join(s.backupDir, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	return name, nil
}

func validateInstance(instance Instance) error {
	if strings.TrimSpace(instance.Name) == "" {
		return errors.New("instance name is required")
	}
	if instance.Engine != EngineSingBox && instance.Engine != EngineXray {
		return errors.New("engine must be sing-box or xray")
	}
	if instance.Scope != ScopeSystem && instance.Scope != ScopeUser {
		return errors.New("scope must be system or user")
	}
	if !unitPattern.MatchString(instance.Unit) {
		return errors.New("invalid systemd unit")
	}
	if !filepath.IsAbs(instance.BinaryPath) || !filepath.IsAbs(instance.ConfigPath) {
		return errors.New("binary and config paths must be absolute")
	}
	switch instance.Engine {
	case EngineSingBox:
		if instance.BinaryPath != "/usr/bin/sing-box" {
			return errors.New("sing-box binary path is not allowed")
		}
	case EngineXray:
		if instance.BinaryPath != "/home/claude/agsbx/xray" && instance.BinaryPath != "/usr/bin/xray" {
			return errors.New("xray binary path is not allowed")
		}
	}
	return nil
}

func parseProperties(output string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		if key, value, ok := strings.Cut(line, "="); ok {
			result[key] = value
		}
	}
	return result
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
