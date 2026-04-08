//go:build !darwin

package tabs

import (
	"fyne.io/fyne/v2/widget"

	wizardmodels "singbox-launcher/ui/wizard/models"
	wizardpresentation "singbox-launcher/ui/wizard/presentation"
	wizardtemplate "singbox-launcher/ui/wizard/template"
)

// maybeTunOffDarwin — только macOS; на других ОС ничего не делает.
func maybeTunOffDarwin(_ *wizardpresentation.WizardPresenter, _ *wizardmodels.WizardModel, _ *wizardtemplate.TemplateData, _ string, _ *widget.Check) bool {
	return false
}
