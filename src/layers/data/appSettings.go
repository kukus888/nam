package data

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidSettingValue = errors.New("invalid app setting value")

type AppSetting struct {
	Id                int64     `json:"id" db:"id"`
	Key               string    `json:"key" db:"key"`
	Value             string    `json:"value" db:"value"`
	ValueType         string    `json:"value_type" db:"value_type"`
	ValidationRegex   *string   `json:"validation_regex" db:"validation_regex"`
	Description       string    `json:"description" db:"description"`
	MinValue          *float64  `json:"min_value" db:"min_value"`
	MaxValue          *float64  `json:"max_value" db:"max_value"`
	AllowedValues     []string  `json:"allowed_values" db:"allowed_values"`
	IsEditable        bool      `json:"is_editable" db:"is_editable"`
	UpdatedBy         *uint64   `json:"updated_by" db:"updated_by"`
	UpdatedByUsername *string   `json:"updated_by_username" db:"updated_by_username"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type DatabaseCapacityGuardConfig struct {
	AllocatedBytes  int64
	WarnPercent     float64
	CriticalPercent float64
	CheckInterval   time.Duration
}

func parseDatabaseCapacityGuardConfigFromLookup(lookup map[string]string) (*DatabaseCapacityGuardConfig, error) {
	requiredKeys := []string{
		"db_capacity_allocated_bytes",
		"db_capacity_warn_percent",
		"db_capacity_critical_percent",
		"db_capacity_check_interval",
	}
	for _, key := range requiredKeys {
		if _, ok := lookup[key]; !ok {
			return nil, fmt.Errorf("missing required capacity setting: %s", key)
		}
	}

	allocatedBytes, err := strconv.ParseInt(lookup["db_capacity_allocated_bytes"], 10, 64)
	if err != nil {
		return nil, err
	}
	warnPercent, err := strconv.ParseFloat(lookup["db_capacity_warn_percent"], 64)
	if err != nil {
		return nil, err
	}
	criticalPercent, err := strconv.ParseFloat(lookup["db_capacity_critical_percent"], 64)
	if err != nil {
		return nil, err
	}
	if criticalPercent <= warnPercent {
		return nil, fmt.Errorf("db_capacity_critical_percent must be greater than db_capacity_warn_percent")
	}
	checkInterval, err := time.ParseDuration(lookup["db_capacity_check_interval"])
	if err != nil {
		return nil, err
	}

	return &DatabaseCapacityGuardConfig{
		AllocatedBytes:  allocatedBytes,
		WarnPercent:     warnPercent,
		CriticalPercent: criticalPercent,
		CheckInterval:   checkInterval,
	}, nil
}

func ValidateDatabaseCapacityGuardSettingsCandidate(pool *pgxpool.Pool, overrides map[string]string) error {
	settings, err := GetAppSettingsByPrefix(pool, "db_capacity_")
	if err != nil {
		return err
	}

	lookup := make(map[string]string, len(*settings))
	for _, setting := range *settings {
		candidateValue := setting.Value
		if overrideValue, ok := overrides[setting.Key]; ok {
			candidateValue = strings.TrimSpace(overrideValue)
		}
		if err := ValidateAppSettingValue(setting, candidateValue); err != nil {
			return err
		}
		lookup[setting.Key] = candidateValue
	}

	_, err = parseDatabaseCapacityGuardConfigFromLookup(lookup)
	return err
}

func GetAppSettingsByPrefix(pool *pgxpool.Pool, prefix string) (*[]AppSetting, error) {
	rows, err := pool.Query(context.Background(), `
		SELECT
			a.id,
			a.key,
			a.value,
			a.value_type,
			a.validation_regex,
			a.description,
			a.min_value,
			a.max_value,
			a.allowed_values,
			a.is_editable,
			a.updated_by,
			u.username AS updated_by_username,
			a.created_at,
			a.updated_at
		FROM app_settings a
		LEFT JOIN "user" u ON u.id = a.updated_by
		WHERE a.key LIKE $1
		ORDER BY a.key ASC;
	`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make([]AppSetting, 0)
	for rows.Next() {
		var s AppSetting
		var updatedBy *int64
		var updatedByUsername *string
		if err := rows.Scan(
			&s.Id,
			&s.Key,
			&s.Value,
			&s.ValueType,
			&s.ValidationRegex,
			&s.Description,
			&s.MinValue,
			&s.MaxValue,
			&s.AllowedValues,
			&s.IsEditable,
			&updatedBy,
			&updatedByUsername,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if updatedBy != nil {
			u := uint64(*updatedBy)
			s.UpdatedBy = &u
		}
		s.UpdatedByUsername = updatedByUsername
		settings = append(settings, s)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return &settings, nil
}

func GetAllAppSettings(pool *pgxpool.Pool) (*[]AppSetting, error) {
	return GetAppSettingsByPrefix(pool, "")
}

func GetAppSettingByKey(pool *pgxpool.Pool, key string) (*AppSetting, error) {
	row := pool.QueryRow(context.Background(), `
		SELECT
			a.id,
			a.key,
			a.value,
			a.value_type,
			a.validation_regex,
			a.description,
			a.min_value,
			a.max_value,
			a.allowed_values,
			a.is_editable,
			a.updated_by,
			u.username AS updated_by_username,
			a.created_at,
			a.updated_at
		FROM app_settings a
		LEFT JOIN "user" u ON u.id = a.updated_by
		WHERE a.key = $1;
	`, key)

	var s AppSetting
	var updatedBy *int64
	var updatedByUsername *string
	if err := row.Scan(
		&s.Id,
		&s.Key,
		&s.Value,
		&s.ValueType,
		&s.ValidationRegex,
		&s.Description,
		&s.MinValue,
		&s.MaxValue,
		&s.AllowedValues,
		&s.IsEditable,
		&updatedBy,
		&updatedByUsername,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if updatedBy != nil {
		u := uint64(*updatedBy)
		s.UpdatedBy = &u
	}
	s.UpdatedByUsername = updatedByUsername

	return &s, nil
}

func UpdateAppSettingValue(pool *pgxpool.Pool, key, value string, updatedBy *uint64) error {
	setting, err := GetAppSettingByKey(pool, key)
	if err != nil {
		return err
	}
	if setting == nil {
		return fmt.Errorf("%w: setting %q does not exist", ErrInvalidSettingValue, key)
	}
	if !setting.IsEditable {
		return fmt.Errorf("%w: setting %q is not editable", ErrInvalidSettingValue, key)
	}

	value = strings.TrimSpace(value)
	if err := ValidateAppSettingValue(*setting, value); err != nil {
		return err
	}

	_, err = pool.Exec(context.Background(), `
		UPDATE app_settings
		SET
			value = $2,
			updated_by = $3,
			updated_at = NOW()
		WHERE key = $1;
	`, key, value, updatedBy)
	return err
}

func ValidateAppSettingValue(setting AppSetting, value string) error {
	if value == "" {
		return fmt.Errorf("%w: %s cannot be empty", ErrInvalidSettingValue, setting.Key)
	}

	if setting.ValidationRegex != nil && *setting.ValidationRegex != "" {
		re, err := regexp.Compile(*setting.ValidationRegex)
		if err != nil {
			return fmt.Errorf("%w: invalid validation regex for key %s", ErrInvalidSettingValue, setting.Key)
		}
		if !re.MatchString(value) {
			return fmt.Errorf("%w: value for %s does not match required format", ErrInvalidSettingValue, setting.Key)
		}
	}

	if len(setting.AllowedValues) > 0 {
		allowed := false
		for _, candidate := range setting.AllowedValues {
			if value == candidate {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("%w: value for %s must be one of %v", ErrInvalidSettingValue, setting.Key, setting.AllowedValues)
		}
	}

	applyRangeCheck := func(parsed float64) error {
		if setting.MinValue != nil && parsed < *setting.MinValue {
			return fmt.Errorf("%w: value for %s is below minimum of %v", ErrInvalidSettingValue, setting.Key, *setting.MinValue)
		}
		if setting.MaxValue != nil && parsed > *setting.MaxValue {
			return fmt.Errorf("%w: value for %s is above maximum of %v", ErrInvalidSettingValue, setting.Key, *setting.MaxValue)
		}
		return nil
	}

	switch setting.ValueType {
	case "int", "size_bytes":
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: value for %s must be an integer", ErrInvalidSettingValue, setting.Key)
		}
		return applyRangeCheck(float64(parsed))
	case "float":
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("%w: value for %s must be a float", ErrInvalidSettingValue, setting.Key)
		}
		return applyRangeCheck(parsed)
	case "bool":
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("%w: value for %s must be true or false", ErrInvalidSettingValue, setting.Key)
		}
		return nil
	case "duration":
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("%w: value for %s must be a valid Go duration (for example 30s, 5m, 1h)", ErrInvalidSettingValue, setting.Key)
		}
		return nil
	case "string":
		return nil
	default:
		return fmt.Errorf("%w: unsupported setting value_type %q", ErrInvalidSettingValue, setting.ValueType)
	}
}

func GetDatabaseCapacityGuardConfig(pool *pgxpool.Pool) (*DatabaseCapacityGuardConfig, error) {
	settings, err := GetAppSettingsByPrefix(pool, "db_capacity_")
	if err != nil {
		return nil, err
	}
	lookup := make(map[string]string, len(*settings))
	for _, s := range *settings {
		lookup[s.Key] = s.Value
	}
	return parseDatabaseCapacityGuardConfigFromLookup(lookup)
}
