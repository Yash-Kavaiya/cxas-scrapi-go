package linter

import (
	"fmt"
	"os"
	"path/filepath"
)

// LintReport summarises all findings from a linting run.
type LintReport struct {
	Files    int
	Errors   int
	Warnings int
	Infos    int
	Results  []LintResult
}

// Add records a finding and increments the appropriate counter.
func (r *LintReport) Add(lr LintResult) {
	r.Results = append(r.Results, lr)
	switch lr.Severity {
	case SeverityError:
		r.Errors++
	case SeverityWarning:
		r.Warnings++
	case SeverityInfo:
		r.Infos++
	}
}

// HasErrors returns true if any error-level findings were recorded.
func (r *LintReport) HasErrors() bool { return r.Errors > 0 }

// Linter orchestrates running all rules against a set of files.
type Linter struct {
	registry *RuleRegistry
	ctx      *LintContext
}

// New creates a Linter using the default rule registry.
func New(appRoot string, severityOverrides map[string]Severity) *Linter {
	return &Linter{
		registry: DefaultRegistry,
		ctx:      &LintContext{AppRoot: appRoot, Severities: severityOverrides},
	}
}

// NewWithRegistry creates a Linter with a custom rule registry.
func NewWithRegistry(registry *RuleRegistry, appRoot string, severityOverrides map[string]Severity) *Linter {
	return &Linter{
		registry: registry,
		ctx:      &LintContext{AppRoot: appRoot, Severities: severityOverrides},
	}
}

// LintFile lints a single file against all registered rules.
func (l *Linter) LintFile(filePath string) ([]LintResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filePath, err)
	}
	content := string(data)

	var results []LintResult
	for _, rule := range l.registry.All() {
		if severityFor(l.ctx, rule.ID(), rule.DefaultSeverity()) == SeverityOff {
			continue
		}
		results = append(results, rule.Check(filePath, content, l.ctx)...)
	}
	return results, nil
}

// LintDirectory lints all .json and .yaml/.yml files in a directory tree.
func (l *Linter) LintDirectory(dir string) (*LintReport, error) {
	report := &LintReport{}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		results, err := l.LintFile(path)
		if err != nil {
			// Non-fatal: record as info and continue.
			report.Add(LintResult{
				File:     path,
				RuleID:   "CXL000",
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("could not read file: %v", err),
			})
			return nil
		}
		report.Files++
		for _, r := range results {
			report.Add(r)
		}
		return nil
	})
	return report, err
}
