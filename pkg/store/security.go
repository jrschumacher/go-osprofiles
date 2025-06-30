package store

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// SecurityLevel defines the security classification for profile fields
type SecurityLevel string

const (
	// SecurityLevelSecure indicates field should be encrypted and stored securely
	SecurityLevelSecure SecurityLevel = "secure"
	// SecurityLevelPlaintext indicates field can be stored in plaintext
	SecurityLevelPlaintext SecurityLevel = "plaintext"
	// SecurityLevelTemporary indicates field should only be held in memory
	SecurityLevelTemporary SecurityLevel = "temporary"
)

// SecurityMode defines the overall security behavior of the storage system
type SecurityMode int

const (
	// SecurityModeKeyring uses hybrid storage with keyring for encryption keys
	SecurityModeKeyring SecurityMode = iota
	// SecurityModeInsecure falls back to plaintext storage with warnings
	SecurityModeInsecure
)

// SecurityConfig holds configuration for the security system
type SecurityConfig struct {
	Mode             SecurityMode
	WarnOnInsecure   bool
	StoreDirectory   string
	ServiceNamespace string
	ProfileKey       string
}

// FieldSecurityInfo contains security metadata for a struct field
type FieldSecurityInfo struct {
	Name      string
	Level     SecurityLevel
	Value     interface{}
	FieldType reflect.Type
}

// SecurityClassification contains the classified fields from a profile
type SecurityClassification struct {
	SecureFields    []FieldSecurityInfo
	PlaintextFields []FieldSecurityInfo
	TemporaryFields []FieldSecurityInfo
}

// DefaultSecurityLevel is used when no security tag is specified
const DefaultSecurityLevel = SecurityLevelPlaintext

// ValidSecurityLevels contains all valid security level values
var ValidSecurityLevels = map[SecurityLevel]bool{
	SecurityLevelSecure:    true,
	SecurityLevelPlaintext: true,
	SecurityLevelTemporary: true,
}

// IsValid checks if a security level is valid
func (sl SecurityLevel) IsValid() bool {
	return ValidSecurityLevels[sl]
}

// String returns the string representation of the security level
func (sl SecurityLevel) String() string {
	return string(sl)
}

// ParseSecurityLevel parses a string into a SecurityLevel
func ParseSecurityLevel(s string) (SecurityLevel, error) {
	level := SecurityLevel(strings.ToLower(strings.TrimSpace(s)))
	if !level.IsValid() {
		return "", fmt.Errorf("%w: invalid security level '%s', must be one of: %v",
			ErrSecurityLevelInvalid, s, getValidSecurityLevelNames())
	}
	return level, nil
}

// getValidSecurityLevelNames returns a list of valid security level names
func getValidSecurityLevelNames() []string {
	var names []string
	for level := range ValidSecurityLevels {
		names = append(names, string(level))
	}
	return names
}

// GetSecurityTag extracts the security tag from a struct field
func GetSecurityTag(field reflect.StructField) SecurityLevel {
	tag := field.Tag.Get("security")
	if tag == "" {
		return DefaultSecurityLevel
	}
	
	level, err := ParseSecurityLevel(tag)
	if err != nil {
		// Invalid tag, use default
		return DefaultSecurityLevel
	}
	
	return level
}

// ClassifyProfile analyzes a profile struct and classifies its fields by security level
func ClassifyProfile(profile interface{}) (*SecurityClassification, error) {
	val := reflect.ValueOf(profile)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: expected struct, got %T", ErrSecurityClassificationFailed, profile)
	}
	
	typ := val.Type()
	classification := &SecurityClassification{
		SecureFields:    make([]FieldSecurityInfo, 0),
		PlaintextFields: make([]FieldSecurityInfo, 0),
		TemporaryFields: make([]FieldSecurityInfo, 0),
	}
	
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		
		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}
		
		level := GetSecurityTag(field)
		fieldInfo := FieldSecurityInfo{
			Name:      field.Name,
			Level:     level,
			Value:     fieldValue.Interface(),
			FieldType: field.Type,
		}
		
		switch level {
		case SecurityLevelSecure:
			classification.SecureFields = append(classification.SecureFields, fieldInfo)
		case SecurityLevelPlaintext:
			classification.PlaintextFields = append(classification.PlaintextFields, fieldInfo)
		case SecurityLevelTemporary:
			classification.TemporaryFields = append(classification.TemporaryFields, fieldInfo)
		}
	}
	
	return classification, nil
}

// ValidateSecurityClassification ensures the security classification is valid
func ValidateSecurityClassification(classification *SecurityClassification) error {
	if classification == nil {
		return fmt.Errorf("%w: classification cannot be nil", ErrSecurityClassificationFailed)
	}
	
	// Check for duplicate field names across security levels
	fieldNames := make(map[string]SecurityLevel)
	
	allFields := [][]FieldSecurityInfo{
		classification.SecureFields,
		classification.PlaintextFields,
		classification.TemporaryFields,
	}
	
	for _, fields := range allFields {
		for _, field := range fields {
			if existingLevel, exists := fieldNames[field.Name]; exists {
				return fmt.Errorf("%w: field '%s' appears in both %s and %s classifications",
					ErrDuplicateFieldClassification, field.Name, existingLevel, field.Level)
			}
			fieldNames[field.Name] = field.Level
		}
	}
	
	return nil
}

// ReconstructProfile creates a new profile instance from classified fields
func ReconstructProfile(originalType reflect.Type, classification *SecurityClassification, includeTemporary bool) (interface{}, error) {
	if originalType.Kind() == reflect.Ptr {
		originalType = originalType.Elem()
	}
	
	if originalType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: expected struct type, got %v", ErrSecurityClassificationFailed, originalType.Kind())
	}
	
	// Create new instance
	newVal := reflect.New(originalType).Elem()
	
	// Set fields from all classifications
	allFields := [][]FieldSecurityInfo{
		classification.SecureFields,
		classification.PlaintextFields,
	}
	
	if includeTemporary {
		allFields = append(allFields, classification.TemporaryFields)
	}
	
	for _, fields := range allFields {
		for _, fieldInfo := range fields {
			field := newVal.FieldByName(fieldInfo.Name)
			if !field.IsValid() || !field.CanSet() {
				continue
			}
			
			fieldVal := reflect.ValueOf(fieldInfo.Value)
			if fieldVal.Type().AssignableTo(field.Type()) {
				field.Set(fieldVal)
			}
		}
	}
	
	return newVal.Interface(), nil
}

// MergeClassifications combines two security classifications, with override taking precedence
func MergeClassifications(base, override *SecurityClassification) *SecurityClassification {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}
	
	merged := &SecurityClassification{
		SecureFields:    make([]FieldSecurityInfo, len(base.SecureFields)),
		PlaintextFields: make([]FieldSecurityInfo, len(base.PlaintextFields)),
		TemporaryFields: make([]FieldSecurityInfo, len(base.TemporaryFields)),
	}
	
	copy(merged.SecureFields, base.SecureFields)
	copy(merged.PlaintextFields, base.PlaintextFields)
	copy(merged.TemporaryFields, base.TemporaryFields)
	
	// Create map of override fields for quick lookup
	overrideFields := make(map[string]FieldSecurityInfo)
	for _, fields := range [][]FieldSecurityInfo{override.SecureFields, override.PlaintextFields, override.TemporaryFields} {
		for _, field := range fields {
			overrideFields[field.Name] = field
		}
	}
	
	// Apply overrides to each classification
	merged.SecureFields = applyFieldOverrides(merged.SecureFields, overrideFields)
	merged.PlaintextFields = applyFieldOverrides(merged.PlaintextFields, overrideFields)
	merged.TemporaryFields = applyFieldOverrides(merged.TemporaryFields, overrideFields)
	
	// Add any new fields from override that don't exist in base
	addNewFields := func(targetFields *[]FieldSecurityInfo, sourceFields []FieldSecurityInfo) {
		baseFieldNames := make(map[string]bool)
		for _, field := range *targetFields {
			baseFieldNames[field.Name] = true
		}
		
		for _, field := range sourceFields {
			if !baseFieldNames[field.Name] {
				*targetFields = append(*targetFields, field)
			}
		}
	}
	
	addNewFields(&merged.SecureFields, override.SecureFields)
	addNewFields(&merged.PlaintextFields, override.PlaintextFields)
	addNewFields(&merged.TemporaryFields, override.TemporaryFields)
	
	return merged
}

// applyFieldOverrides applies override values to a slice of fields
func applyFieldOverrides(fields []FieldSecurityInfo, overrides map[string]FieldSecurityInfo) []FieldSecurityInfo {
	for i, field := range fields {
		if override, exists := overrides[field.Name]; exists {
			fields[i] = override
		}
	}
	return fields
}

// GetFieldValue safely extracts a field value from a profile using reflection
func GetFieldValue(profile interface{}, fieldName string) (interface{}, error) {
	val := reflect.ValueOf(profile)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: expected struct, got %T", ErrSecurityClassificationFailed, profile)
	}
	
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("%w: field '%s' not found", ErrSecurityClassificationFailed, fieldName)
	}
	
	if !field.CanInterface() {
		return nil, fmt.Errorf("%w: field '%s' is not accessible", ErrSecurityClassificationFailed, fieldName)
	}
	
	return field.Interface(), nil
}

// SetFieldValue safely sets a field value in a profile using reflection
func SetFieldValue(profile interface{}, fieldName string, value interface{}) error {
	val := reflect.ValueOf(profile)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("%w: profile must be a pointer to set fields", ErrSecurityClassificationFailed)
	}
	
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("%w: expected struct, got %T", ErrSecurityClassificationFailed, profile)
	}
	
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("%w: field '%s' not found", ErrSecurityClassificationFailed, fieldName)
	}
	
	if !field.CanSet() {
		return fmt.Errorf("%w: field '%s' cannot be set", ErrSecurityClassificationFailed, fieldName)
	}
	
	fieldVal := reflect.ValueOf(value)
	if !fieldVal.Type().AssignableTo(field.Type()) {
		return fmt.Errorf("%w: cannot assign %T to field '%s' of type %v", 
			ErrSecurityClassificationFailed, value, fieldName, field.Type())
	}
	
	field.Set(fieldVal)
	return nil
}

// CopyProfileStructure creates a new profile instance with the same structure but zero values
func CopyProfileStructure(profile interface{}) (interface{}, error) {
	val := reflect.ValueOf(profile)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: expected struct, got %T", ErrSecurityClassificationFailed, profile)
	}
	
	typ := val.Type()
	newVal := reflect.New(typ).Elem()
	
	return newVal.Interface(), nil
}

// SplitData represents the split data for different security levels
type SplitData struct {
	SecureData    map[string]interface{} `json:"secure_data,omitempty"`
	PlaintextData map[string]interface{} `json:"plaintext_data,omitempty"`
	TemporaryData map[string]interface{} `json:"-"` // Never serialized
}

// SplitProfileData splits a profile into separate data maps based on security classification
func SplitProfileData(profile interface{}) (*SplitData, error) {
	classification, err := ClassifyProfile(profile)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to classify profile: %v", ErrSecurityClassificationFailed, err)
	}
	
	if err := ValidateSecurityClassification(classification); err != nil {
		return nil, fmt.Errorf("%w: invalid classification: %v", ErrSecurityValidationFailed, err)
	}
	
	split := &SplitData{
		SecureData:    make(map[string]interface{}),
		PlaintextData: make(map[string]interface{}),
		TemporaryData: make(map[string]interface{}),
	}
	
	// Split fields into appropriate data maps
	for _, field := range classification.SecureFields {
		split.SecureData[field.Name] = field.Value
	}
	
	for _, field := range classification.PlaintextFields {
		split.PlaintextData[field.Name] = field.Value
	}
	
	for _, field := range classification.TemporaryFields {
		split.TemporaryData[field.Name] = field.Value
	}
	
	return split, nil
}

// CombineSplitData combines split data back into a profile of the specified type
func CombineSplitData(profileType reflect.Type, split *SplitData, includeTemporary bool) (interface{}, error) {
	if profileType.Kind() == reflect.Ptr {
		profileType = profileType.Elem()
	}
	
	if profileType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: expected struct type, got %v", ErrSecurityClassificationFailed, profileType.Kind())
	}
	
	// Create new instance
	newVal := reflect.New(profileType).Elem()
	
	// Set fields from split data
	allData := []map[string]interface{}{
		split.SecureData,
		split.PlaintextData,
	}
	
	if includeTemporary {
		allData = append(allData, split.TemporaryData)
	}
	
	for _, dataMap := range allData {
		for fieldName, value := range dataMap {
			field := newVal.FieldByName(fieldName)
			if !field.IsValid() || !field.CanSet() {
				continue
			}
			
			fieldVal := reflect.ValueOf(value)
			if fieldVal.Type().AssignableTo(field.Type()) {
				field.Set(fieldVal)
			}
		}
	}
	
	return newVal.Interface(), nil
}

// SerializeSecureData converts secure data to JSON bytes for encryption
func SerializeSecureData(split *SplitData) ([]byte, error) {
	if split == nil || len(split.SecureData) == 0 {
		return []byte("{}"), nil
	}
	
	data, err := json.Marshal(split.SecureData)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to serialize secure data: %v", ErrSecurityClassificationFailed, err)
	}
	
	return data, nil
}

// SerializePlaintextData converts plaintext data to JSON bytes
func SerializePlaintextData(split *SplitData) ([]byte, error) {
	if split == nil || len(split.PlaintextData) == 0 {
		return []byte("{}"), nil
	}
	
	data, err := json.Marshal(split.PlaintextData)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to serialize plaintext data: %v", ErrSecurityClassificationFailed, err)
	}
	
	return data, nil
}

// DeserializeSplitData reconstructs split data from JSON bytes
func DeserializeSplitData(secureBytes, plaintextBytes []byte) (*SplitData, error) {
	split := &SplitData{
		SecureData:    make(map[string]interface{}),
		PlaintextData: make(map[string]interface{}),
		TemporaryData: make(map[string]interface{}),
	}
	
	if len(secureBytes) > 0 {
		if err := json.Unmarshal(secureBytes, &split.SecureData); err != nil {
			return nil, fmt.Errorf("%w: failed to deserialize secure data: %v", ErrSecurityClassificationFailed, err)
		}
	}
	
	if len(plaintextBytes) > 0 {
		if err := json.Unmarshal(plaintextBytes, &split.PlaintextData); err != nil {
			return nil, fmt.Errorf("%w: failed to deserialize plaintext data: %v", ErrSecurityClassificationFailed, err)
		}
	}
	
	return split, nil
}

// AddTemporaryData adds temporary field data to split data (in-memory only)
func (sd *SplitData) AddTemporaryData(fieldName string, value interface{}) {
	if sd.TemporaryData == nil {
		sd.TemporaryData = make(map[string]interface{})
	}
	sd.TemporaryData[fieldName] = value
}

// RemoveTemporaryData removes temporary field data
func (sd *SplitData) RemoveTemporaryData(fieldName string) {
	if sd.TemporaryData != nil {
		delete(sd.TemporaryData, fieldName)
	}
}

// ClearTemporaryData removes all temporary data
func (sd *SplitData) ClearTemporaryData() {
	sd.TemporaryData = make(map[string]interface{})
}

// HasSecureData returns true if there is secure data present
func (sd *SplitData) HasSecureData() bool {
	return sd != nil && len(sd.SecureData) > 0
}

// HasPlaintextData returns true if there is plaintext data present
func (sd *SplitData) HasPlaintextData() bool {
	return sd != nil && len(sd.PlaintextData) > 0
}

// HasTemporaryData returns true if there is temporary data present
func (sd *SplitData) HasTemporaryData() bool {
	return sd != nil && len(sd.TemporaryData) > 0
}

// ValidateProfileSecurity validates the security configuration of a profile struct
func ValidateProfileSecurity(profile interface{}) error {
	val := reflect.ValueOf(profile)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("%w: expected struct, got %T", ErrSecurityValidationFailed, profile)
	}
	
	typ := val.Type()
	seenFields := make(map[string]SecurityLevel)
	
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		
		// Skip unexported fields
		if !val.Field(i).CanInterface() {
			continue
		}
		
		level := GetSecurityTag(field)
		
		// Check for invalid security tags
		if !level.IsValid() {
			return fmt.Errorf("%w: field '%s' has invalid security level '%s'", 
				ErrSecurityValidationFailed, field.Name, level)
		}
		
		// Check for duplicate field names (shouldn't happen but good to validate)
		if existingLevel, exists := seenFields[field.Name]; exists {
			return fmt.Errorf("%w: duplicate field name '%s' with different security levels: %s vs %s",
				ErrDuplicateFieldClassification, field.Name, existingLevel, level)
		}
		
		seenFields[field.Name] = level
	}
	
	return nil
}

// ValidateSecurityMode validates that the security mode is compatible with available features
func ValidateSecurityMode(mode SecurityMode, keyringAvailable bool) error {
	switch mode {
	case SecurityModeKeyring:
		if !keyringAvailable {
			return fmt.Errorf("%w: keyring mode requested but keyring is not available", ErrKeyringUnavailable)
		}
		return nil
	case SecurityModeInsecure:
		// Always valid, but should warn
		return nil
	default:
		return fmt.Errorf("%w: invalid security mode %d", ErrSecurityModeUnavailable, mode)
	}
}

// ValidateSecurityConfig validates a complete security configuration
func ValidateSecurityConfig(config *SecurityConfig) error {
	if config == nil {
		return fmt.Errorf("%w: security config cannot be nil", ErrSecurityValidationFailed)
	}
	
	if config.ServiceNamespace == "" {
		return fmt.Errorf("%w: service namespace cannot be empty", ErrSecurityValidationFailed)
	}
	
	if config.ProfileKey == "" {
		return fmt.Errorf("%w: profile key cannot be empty", ErrSecurityValidationFailed)
	}
	
	// Validate namespace and key format
	if err := ValidateNamespaceKey(config.ServiceNamespace, config.ProfileKey); err != nil {
		return fmt.Errorf("%w: invalid namespace/key: %v", ErrSecurityValidationFailed, err)
	}
	
	return nil
}

// CheckSecurityCompliance checks if a profile meets security requirements
func CheckSecurityCompliance(profile interface{}, requireSecureFields bool) error {
	classification, err := ClassifyProfile(profile)
	if err != nil {
		return fmt.Errorf("%w: failed to classify profile: %v", ErrSecurityValidationFailed, err)
	}
	
	if requireSecureFields && len(classification.SecureFields) == 0 {
		return fmt.Errorf("%w: profile must contain at least one secure field", ErrSecurityValidationFailed)
	}
	
	// Check that sensitive field names don't appear in plaintext
	sensitiveKeywords := []string{"password", "secret", "key", "token", "credential", "auth"}
	
	for _, field := range classification.PlaintextFields {
		fieldNameLower := strings.ToLower(field.Name)
		for _, keyword := range sensitiveKeywords {
			if strings.Contains(fieldNameLower, keyword) && field.Level != SecurityLevelSecure {
				return fmt.Errorf("%w: field '%s' appears to contain sensitive data but is not marked as secure",
					ErrSecurityValidationFailed, field.Name)
			}
		}
	}
	
	return nil
}

// ValidateFieldSecurityTag validates a specific field's security tag
func ValidateFieldSecurityTag(field reflect.StructField) error {
	tag := field.Tag.Get("security")
	if tag == "" {
		// No tag is valid (uses default)
		return nil
	}
	
	_, err := ParseSecurityLevel(tag)
	if err != nil {
		return fmt.Errorf("%w: field '%s' has invalid security tag: %v", ErrSecurityValidationFailed, field.Name, err)
	}
	
	// Additional validation rules could go here
	// For example, certain field types might have restrictions
	
	return nil
}

// GetSecurityReport generates a security report for a profile
type SecurityReport struct {
	SecureFieldCount    int      `json:"secure_field_count"`
	PlaintextFieldCount int      `json:"plaintext_field_count"`
	TemporaryFieldCount int      `json:"temporary_field_count"`
	SecurityWarnings    []string `json:"security_warnings,omitempty"`
	SecurityErrors      []string `json:"security_errors,omitempty"`
}

// GenerateSecurityReport creates a comprehensive security report for a profile
func GenerateSecurityReport(profile interface{}) (*SecurityReport, error) {
	classification, err := ClassifyProfile(profile)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to classify profile: %v", ErrSecurityValidationFailed, err)
	}
	
	report := &SecurityReport{
		SecureFieldCount:    len(classification.SecureFields),
		PlaintextFieldCount: len(classification.PlaintextFields),
		TemporaryFieldCount: len(classification.TemporaryFields),
		SecurityWarnings:    make([]string, 0),
		SecurityErrors:      make([]string, 0),
	}
	
	// Check for potential security issues
	sensitiveKeywords := []string{"password", "secret", "key", "token", "credential", "auth"}
	
	for _, field := range classification.PlaintextFields {
		fieldNameLower := strings.ToLower(field.Name)
		for _, keyword := range sensitiveKeywords {
			if strings.Contains(fieldNameLower, keyword) {
				report.SecurityWarnings = append(report.SecurityWarnings,
					fmt.Sprintf("Field '%s' contains sensitive keyword '%s' but is stored in plaintext", field.Name, keyword))
			}
		}
	}
	
	// Check if profile has no secure fields
	if len(classification.SecureFields) == 0 {
		report.SecurityWarnings = append(report.SecurityWarnings,
			"Profile contains no secure fields - all data will be stored in plaintext")
	}
	
	// Validate field structure
	if err := ValidateSecurityClassification(classification); err != nil {
		report.SecurityErrors = append(report.SecurityErrors, err.Error())
	}
	
	return report, nil
}