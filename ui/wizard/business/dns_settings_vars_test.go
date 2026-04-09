package business

import (
	"testing"

	wizardmodels "singbox-launcher/ui/wizard/models"
	wizardtemplate "singbox-launcher/ui/wizard/template"
)

func TestMigrateDNSScalarsFromPersistedToSettingsVars_Idempotent(t *testing.T) {
	vars := []wizardtemplate.TemplateVar{
		{Name: wizardmodels.VarDNSStrategy, Type: "custom"},
		{Name: wizardmodels.VarDNSFinal, Type: "custom"},
		{Name: wizardmodels.VarDNSIndependentCache, Type: "custom"},
		{Name: wizardmodels.VarDNSDefaultDomainResolver, Type: "custom"},
	}
	st := map[string]string{}
	p := &wizardmodels.PersistedDNSState{
		Strategy:              "prefer_ipv6",
		Final:                 "google_doh",
		DefaultDomainResolver: "direct_dns_resolver",
	}
	b := true
	p.IndependentCache = &b

	MigrateDNSScalarsFromPersistedToSettingsVars(p, st, vars)
	if st[wizardmodels.VarDNSStrategy] != "prefer_ipv6" {
		t.Fatalf("strategy: %v", st)
	}
	MigrateDNSScalarsFromPersistedToSettingsVars(p, st, vars)
	if st[wizardmodels.VarDNSStrategy] != "prefer_ipv6" {
		t.Fatalf("second migrate should not overwrite")
	}
}

func TestMigrateDNSScalarsFromPersisted_DoesNotOverwriteExistingVar(t *testing.T) {
	vars := []wizardtemplate.TemplateVar{{Name: wizardmodels.VarDNSStrategy, Type: "custom"}}
	st := map[string]string{wizardmodels.VarDNSStrategy: "ipv4_only"}
	p := &wizardmodels.PersistedDNSState{Strategy: "prefer_ipv6"}
	MigrateDNSScalarsFromPersistedToSettingsVars(p, st, vars)
	if st[wizardmodels.VarDNSStrategy] != "ipv4_only" {
		t.Fatalf("got %q", st[wizardmodels.VarDNSStrategy])
	}
}
