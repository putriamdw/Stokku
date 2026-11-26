package model

// RuleError menandai pelanggaran aturan bisnis
type RuleError struct{ Message string }

func (e *RuleError) Error() string { return e.Message }