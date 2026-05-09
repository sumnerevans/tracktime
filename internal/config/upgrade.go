package config

import (
	_ "embed"
	"strings"

	"gopkg.in/yaml.v3"

	up "go.mau.fi/util/configupgrade"
)

//go:embed base-config.yaml
var baseConfig string

// Upgrader migrates tracktimerc files from the Python flat format to the Go
// nested format. Running it on an already-migrated file is a no-op.
var Upgrader = &up.StructUpgrader{
	SimpleUpgrader: doUpgrade,
	Base:           baseConfig,
}

func doUpgrade(helper up.Helper) {
	// Fields unchanged between Python and Go
	helper.Copy(up.Str, "directory")
	helper.Copy(up.Str, "editor")
	helper.Copy(up.Str, "typst_path")
	helper.Copy(up.Int, "item_cache_ttl_days")

	// editor_args: Python = comma-separated string, Go = YAML list
	if val, ok := helper.Get(up.Str, "editor_args"); ok {
		setStrList(helper, strings.Split(val, ","), "editor_args")
	} else {
		helper.Copy(up.List, "editor_args")
	}

	// Service sub-sections — same paths in both formats
	helper.Copy(up.Map, "github")
	helper.Copy(up.Map, "gitlab")
	helper.Copy(up.Map, "sourcehut")
	helper.Copy(up.Map, "linear")
	helper.Copy(up.Map, "logging")

	// sync section
	helper.Copy(up.Map, "sync")
	// Python: sync_time: bool (flat) → Go: sync.enable
	if val, ok := helper.Get(up.Bool, "sync_time"); ok {
		helper.Set(up.Bool, val, "sync", "enable")
	}

	// reporting section — copy whole map if already in Go format
	helper.Copy(up.Map, "reporting")

	// Migrate Python flat scalar keys → reporting.*
	if val, ok := helper.Get(up.Str, "fullname"); ok {
		helper.Set(up.Str, val, "reporting", "fullname")
	}
	if val, ok := helper.Get(up.Int, "day_worked_min_threshold"); ok {
		helper.Set(up.Int, val, "reporting", "day_worked_min_threshold")
	}
	if val, ok := helper.Get(up.Bool, "report_statistics"); ok {
		helper.Set(up.Bool, val, "reporting", "report_statistics")
	}

	// Migrate Python flat map keys → reporting.*
	for _, key := range []string{"project_rates", "customer_rates", "customer_aliases", "customer_addresses"} {
		if node := helper.GetNode(key); node != nil && node.Map != nil {
			helper.SetMap(node.Map, "reporting", key)
		}
	}
}

// setStrList writes values as a YAML sequence into the base node at path.
func setStrList(helper up.Helper, values []string, path ...string) {
	node := helper.GetBaseNode(path...)
	if node == nil || node.Node == nil {
		return
	}
	node.Kind = yaml.SequenceNode
	node.Tag = "!!seq"
	node.Value = ""
	node.Content = nil
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		node.Content = append(node.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: v,
		})
	}
}
