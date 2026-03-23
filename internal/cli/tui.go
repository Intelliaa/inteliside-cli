package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/backup"
	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/Intelliaa/inteliside-cli/internal/deps"
	"github.com/Intelliaa/inteliside-cli/internal/model"
	"github.com/Intelliaa/inteliside-cli/internal/pipeline"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUI states
type tuiState int

const (
	stateWelcome tuiState = iota
	statePresetSelect
	statePluginSelect
	stateDepReview
	stateSecrets
	stateConfirm
	stateInstalling
	stateDone
	stateError
)

// secretField tracks which secret is being collected
type secretField struct {
	key         string
	label       string
	hint        string
	value       string
	masked      bool
	required    bool
	forPlugins  []string // which plugin IDs need this
}

type tuiModel struct {
	state           tuiState
	presets         []model.Preset
	plugins         []model.Plugin
	cursor          int
	selectedPreset  string
	selectedPlugins map[string]bool
	steps           []model.Step
	result          *pipeline.Result
	err             error
	width           int
	height          int

	// dep review
	resolvedSteps []model.Step

	// secrets
	secretFields  []secretField
	secretCursor  int
	secretEditing bool
	secrets       map[string]string

	// progress tracking
	currentStep   int
	stepLogs      []string
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 2)

	secretInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7C3AED")).
				Background(lipgloss.Color("#1F2937")).
				Padding(0, 1)

	stepOkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	stepPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280"))

	stepActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)
)

const asciiLogo = `
    ██╗███╗   ██╗████████╗███████╗██╗     ██╗███████╗██╗██████╗ ███████╗
    ██║████╗  ██║╚══██╔══╝██╔════╝██║     ██║██╔════╝██║██╔══██╗██╔════╝
    ██║██╔██╗ ██║   ██║   █████╗  ██║     ██║███████╗██║██║  ██║█████╗
    ██║██║╚██╗██║   ██║   ██╔══╝  ██║     ██║╚════██║██║██║  ██║██╔══╝
    ██║██║ ╚████║   ██║   ███████╗███████╗██║███████║██║██████╔╝███████╗
    ╚═╝╚═╝  ╚═══╝   ╚═╝   ╚══════╝╚══════╝╚═╝╚══════╝╚═╝╚═════╝ ╚══════╝`

func runTUI() error {
	m := tuiModel{
		state:           stateWelcome,
		presets:         catalog.AllPresets(),
		plugins:         catalog.AllPlugins(),
		selectedPlugins: make(map[string]bool),
		secrets:         make(map[string]string),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	fm := finalModel.(tuiModel)
	if fm.err != nil {
		return fm.err
	}
	return nil
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Secret editing mode captures all keys
		if m.state == stateSecrets && m.secretEditing {
			return m.updateSecretInput(msg)
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.state == stateSecrets && m.secretEditing {
				break
			}
			return m, tea.Quit
		case "esc":
			if m.state == stateSecrets && m.secretEditing {
				m.secretEditing = false
				return m, nil
			}
			// Go back one state
			return m.goBack(), nil

		case "up", "k":
			if m.state != stateSecrets || !m.secretEditing {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if m.state != stateSecrets || !m.secretEditing {
				m.moveDown()
			}

		case " ":
			if m.state == statePluginSelect {
				p := m.plugins[m.cursor]
				m.selectedPlugins[p.ID] = !m.selectedPlugins[p.ID]
			}

		case "s":
			if m.state == stateSecrets && !m.secretEditing {
				// Skip secrets, go to confirm
				m.state = stateConfirm
				return m, nil
			}

		case "enter":
			return m.handleEnter()
		}

	case installDoneMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
			m.result = msg.result
		} else {
			m.state = stateDone
			m.result = msg.result
		}

	case stepProgressMsg:
		m.stepLogs = append(m.stepLogs, msg.log)
		m.currentStep = msg.step
	}

	return m, nil
}

func (m tuiModel) moveDown() {
	switch m.state {
	case statePresetSelect:
		if m.cursor < len(m.presets)-1 {
			m.cursor++
		}
	case statePluginSelect:
		if m.cursor < len(m.plugins)-1 {
			m.cursor++
		}
	case stateSecrets:
		if m.cursor < len(m.secretFields)-1 {
			m.cursor++
		}
	}
}

func (m tuiModel) goBack() tuiModel {
	switch m.state {
	case statePresetSelect:
		m.state = stateWelcome
	case statePluginSelect:
		m.state = statePresetSelect
		m.cursor = 0
	case stateDepReview:
		if m.selectedPreset == "custom" {
			m.state = statePluginSelect
		} else {
			m.state = statePresetSelect
		}
		m.cursor = 0
	case stateSecrets:
		m.state = stateDepReview
		m.cursor = 0
	case stateConfirm:
		if len(m.secretFields) > 0 {
			m.state = stateSecrets
		} else {
			m.state = stateDepReview
		}
		m.cursor = 0
	}
	return m
}

func (m tuiModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateWelcome:
		m.state = statePresetSelect
		m.cursor = 0

	case statePresetSelect:
		preset := m.presets[m.cursor]
		m.selectedPreset = preset.ID
		if preset.ID == "custom" {
			m.state = statePluginSelect
			m.cursor = 0
		} else {
			for _, pid := range preset.PluginIDs {
				m.selectedPlugins[pid] = true
			}
			m = m.resolveAndReview()
		}

	case statePluginSelect:
		hasSelected := false
		for _, v := range m.selectedPlugins {
			if v {
				hasSelected = true
				break
			}
		}
		if !hasSelected {
			return m, nil // need at least one
		}
		m = m.resolveAndReview()

	case stateDepReview:
		m.secretFields = m.collectSecretFields()
		if len(m.secretFields) > 0 {
			m.state = stateSecrets
			m.cursor = 0
		} else {
			m.state = stateConfirm
		}

	case stateSecrets:
		if !m.secretEditing {
			m.secretEditing = true
		}

	case stateConfirm:
		m.state = stateInstalling
		m.stepLogs = nil
		m.currentStep = 0
		return m, m.runInstall()

	case stateDone, stateError:
		return m, tea.Quit
	}

	return m, nil
}

func (m tuiModel) updateSecretInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	idx := m.cursor

	switch key {
	case "enter":
		m.secretEditing = false
		m.secrets[m.secretFields[idx].key] = m.secretFields[idx].value
		// Move to next field or to confirm
		if idx < len(m.secretFields)-1 {
			m.cursor++
		}
		return m, nil
	case "esc":
		m.secretEditing = false
		return m, nil
	case "backspace":
		if len(m.secretFields[idx].value) > 0 {
			m.secretFields[idx].value = m.secretFields[idx].value[:len(m.secretFields[idx].value)-1]
		}
		return m, nil
	case "tab":
		m.secretEditing = false
		m.secrets[m.secretFields[idx].key] = m.secretFields[idx].value
		if idx < len(m.secretFields)-1 {
			m.cursor++
			m.secretEditing = true
		}
		return m, nil
	default:
		if len(key) == 1 {
			m.secretFields[idx].value += key
		}
		return m, nil
	}
}

func (m tuiModel) resolveAndReview() tuiModel {
	var pluginIDs []string
	for id, selected := range m.selectedPlugins {
		if selected {
			pluginIDs = append(pluginIDs, id)
		}
	}
	steps, err := deps.Resolve(pluginIDs)
	if err != nil {
		m.state = stateError
		m.err = err
		return m
	}
	m.resolvedSteps = steps
	m.state = stateDepReview
	m.cursor = 0
	return m
}

func (m tuiModel) collectSecretFields() []secretField {
	var fields []secretField
	needsFigma := false
	needsN8n := false
	needsEngram := false

	for id, selected := range m.selectedPlugins {
		if !selected {
			continue
		}
		switch id {
		case "ux-studio":
			needsFigma = true
		case "n8n-studio":
			needsN8n = true
		case "atl-inteliside", "sdd-intake", "sdd-legacy":
			needsEngram = true
		}
	}

	if needsFigma {
		// Check if already configured
		ok, _, _ := catalog.DependencyByID("figma-mcp").CheckFn()
		if !ok {
			fields = append(fields, secretField{
				key:        "figma_token",
				label:      "Figma Access Token",
				hint:       "Figma → Settings → Personal access tokens",
				masked:     true,
				required:   true,
				forPlugins: []string{"ux-studio"},
			})
		}
	}

	if needsN8n {
		// n8n uses shared default, but allow override
		fields = append(fields, secretField{
			key:        "n8n_mcp_url",
			label:      "n8n MCP URL (Enter para usar default)",
			hint:       "Default: https://n8n-mcp.codetrain.cloud/mcp",
			masked:     false,
			required:   false,
			forPlugins: []string{"n8n-studio"},
		})
		fields = append(fields, secretField{
			key:        "n8n_mcp_token",
			label:      "n8n MCP Token (Enter para usar default)",
			hint:       "Default: token compartido de Inteliside",
			masked:     true,
			required:   false,
			forPlugins: []string{"n8n-studio"},
		})
	}

	if needsEngram {
		fields = append(fields, secretField{
			key:        "engram_project",
			label:      "Nombre del proyecto Engram",
			hint:       "Ej: mi-proyecto-dev",
			masked:     false,
			required:   false,
			forPlugins: []string{"atl-inteliside"},
		})
		fields = append(fields, secretField{
			key:        "github_owner",
			label:      "GitHub owner (org o usuario)",
			hint:       "Ej: Intelliaa",
			masked:     false,
			required:   false,
			forPlugins: []string{"atl-inteliside"},
		})
		fields = append(fields, secretField{
			key:        "github_repo",
			label:      "GitHub repo name",
			hint:       "Ej: mi-proyecto",
			masked:     false,
			required:   false,
			forPlugins: []string{"atl-inteliside"},
		})
	}

	return fields
}

type installDoneMsg struct {
	result *pipeline.Result
	err    error
}

type stepProgressMsg struct {
	step int
	log  string
}

func (m tuiModel) runInstall() tea.Cmd {
	return func() tea.Msg {
		var pluginIDs []string
		for id, selected := range m.selectedPlugins {
			if selected {
				pluginIDs = append(pluginIDs, id)
			}
		}

		steps, err := deps.Resolve(pluginIDs)
		if err != nil {
			return installDoneMsg{err: err}
		}

		projectDir, _ := os.Getwd()
		targets := backup.TargetFiles(projectDir)
		backup.Create(targets)

		ctx := model.NewInstallContext(projectDir)
		// Inject collected secrets
		for k, v := range m.secrets {
			ctx.Secrets[k] = v
		}

		result := pipeline.Run(steps, ctx)

		if result.Error != nil {
			return installDoneMsg{result: &result, err: result.Error}
		}
		return installDoneMsg{result: &result}
	}
}

// ── View ──

func (m tuiModel) View() string {
	var b strings.Builder

	switch m.state {
	case stateWelcome:
		b.WriteString(m.viewWelcome())
	case statePresetSelect:
		b.WriteString(m.viewPresetSelect())
	case statePluginSelect:
		b.WriteString(m.viewPluginSelect())
	case stateDepReview:
		b.WriteString(m.viewDepReview())
	case stateSecrets:
		b.WriteString(m.viewSecrets())
	case stateConfirm:
		b.WriteString(m.viewConfirm())
	case stateInstalling:
		b.WriteString(m.viewInstalling())
	case stateDone:
		b.WriteString(m.viewDone())
	case stateError:
		b.WriteString(m.viewError())
	}

	return b.String()
}

func (m tuiModel) viewWelcome() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(brandStyle.Render(asciiLogo))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("                    Plugin Marketplace CLI v" + version))
	b.WriteString("\n\n")
	b.WriteString(normalStyle.Render("    Un comando. Cualquier rol. Cualquier plataforma."))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("    Presiona Enter para comenzar • q para salir"))
	return b.String()
}

func (m tuiModel) viewPresetSelect() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Selecciona tu rol:"))
	b.WriteString("\n\n")

	for i, p := range m.presets {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}

		icon := roleIcon(p.ID)
		b.WriteString(fmt.Sprintf("  %s%s %s\n", cursor, icon, style.Render(p.Name)))
		b.WriteString(fmt.Sprintf("      %s\n", dimStyle.Render(p.Description)))
		if p.PluginIDs != nil {
			b.WriteString(fmt.Sprintf("      %s\n", dimStyle.Render("Plugins: "+strings.Join(p.PluginIDs, ", "))))
		}
		b.WriteString("\n")
	}

	b.WriteString(dimStyle.Render("  ↑↓ navegar • Enter seleccionar • Esc atrás • q salir"))
	return b.String()
}

func (m tuiModel) viewPluginSelect() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Selecciona plugins:"))
	b.WriteString("\n\n")

	count := 0
	for _, v := range m.selectedPlugins {
		if v {
			count++
		}
	}

	for i, p := range m.plugins {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}
		check := "[ ]"
		checkStyle := dimStyle
		if m.selectedPlugins[p.ID] {
			check = "[●]"
			checkStyle = selectedStyle
		}
		style := normalStyle
		if i == m.cursor {
			style = selectedStyle
		}
		b.WriteString(fmt.Sprintf("  %s%s %s\n", cursor, checkStyle.Render(check), style.Render(p.Name)))
		b.WriteString(fmt.Sprintf("        %s\n", dimStyle.Render(p.Description)))
	}

	b.WriteString("\n")
	if count > 0 {
		b.WriteString(fmt.Sprintf("  %s seleccionados\n\n", selectedStyle.Render(fmt.Sprintf("%d plugins", count))))
	}
	b.WriteString(dimStyle.Render("  ↑↓ navegar • Espacio marcar • Enter confirmar • Esc atrás"))
	return b.String()
}

func (m tuiModel) viewDepReview() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Plan de instalación:"))
	b.WriteString("\n\n")

	// Separate deps and plugins
	var depSteps, pluginSteps []model.Step
	for _, s := range m.resolvedSteps {
		if strings.HasPrefix(s.ID, "plugin:") {
			pluginSteps = append(pluginSteps, s)
		} else {
			depSteps = append(depSteps, s)
		}
	}

	if len(depSteps) > 0 {
		b.WriteString(normalStyle.Render("  Dependencias a verificar/instalar:"))
		b.WriteString("\n")
		for i, s := range depSteps {
			// Check current status
			dep := catalog.DependencyByID(s.ID)
			icon := "○"
			status := ""
			if dep != nil && dep.CheckFn != nil {
				ok, detail, _ := dep.CheckFn()
				if ok {
					icon = successStyle.Render("✓")
					status = dimStyle.Render(" (ya instalado)")
				} else {
					icon = warnStyle.Render("→")
					status = dimStyle.Render(" (" + detail + ")")
				}
			}
			b.WriteString(fmt.Sprintf("    %d. %s %s%s\n", i+1, icon, s.Name, status))
		}
		b.WriteString("\n")
	}

	if len(pluginSteps) > 0 {
		b.WriteString(normalStyle.Render("  Plugins:"))
		b.WriteString("\n")
		for _, s := range pluginSteps {
			b.WriteString(fmt.Sprintf("    ● %s\n", selectedStyle.Render(s.Name)))
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("  Total: %s\n\n", normalStyle.Render(fmt.Sprintf("%d pasos", len(m.resolvedSteps)))))
	b.WriteString(dimStyle.Render("  Enter continuar • Esc atrás • q salir"))
	return b.String()
}

func (m tuiModel) viewSecrets() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Configuración de credenciales:"))
	b.WriteString("\n\n")

	for i, f := range m.secretFields {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}

		style := normalStyle
		if i == m.cursor {
			style = selectedStyle
		}

		reqTag := ""
		if f.required {
			reqTag = warnStyle.Render(" (requerido)")
		} else {
			reqTag = dimStyle.Render(" (opcional)")
		}

		b.WriteString(fmt.Sprintf("  %s%s%s\n", cursor, style.Render(f.label), reqTag))
		b.WriteString(fmt.Sprintf("      %s\n", dimStyle.Render(f.hint)))

		// Show input field
		val := f.value
		if f.masked && len(val) > 0 {
			val = strings.Repeat("•", len(val))
		}
		if i == m.cursor && m.secretEditing {
			b.WriteString(fmt.Sprintf("      %s%s\n", secretInputStyle.Render(val), selectedStyle.Render("█")))
		} else if val != "" {
			b.WriteString(fmt.Sprintf("      %s\n", dimStyle.Render(val)))
		} else {
			b.WriteString(fmt.Sprintf("      %s\n", dimStyle.Render("(vacío)")))
		}
		b.WriteString("\n")
	}

	if m.secretEditing {
		b.WriteString(dimStyle.Render("  Escribe el valor • Enter guardar • Tab siguiente • Esc cancelar"))
	} else {
		b.WriteString(dimStyle.Render("  ↑↓ navegar • Enter editar • Esc atrás"))

		// Show skip option
		allOptional := true
		for _, f := range m.secretFields {
			if f.required && f.value == "" {
				allOptional = false
				break
			}
		}
		if allOptional {
			b.WriteString("\n")
			b.WriteString(dimStyle.Render("  's' para saltar y usar defaults"))
		}
	}
	return b.String()
}

func (m tuiModel) viewConfirm() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Listo para instalar:"))
	b.WriteString("\n\n")

	// Plugins
	b.WriteString(normalStyle.Render("  Plugins:"))
	b.WriteString("\n")
	for id, selected := range m.selectedPlugins {
		if selected {
			p := catalog.PluginByID(id)
			if p != nil {
				b.WriteString(fmt.Sprintf("    ● %s — %s\n", selectedStyle.Render(p.Name), dimStyle.Render(p.Description)))
			}
		}
	}
	b.WriteString("\n")

	// Secrets configured
	configuredSecrets := 0
	for _, f := range m.secretFields {
		if f.value != "" {
			configuredSecrets++
		}
	}
	if len(m.secretFields) > 0 {
		b.WriteString(fmt.Sprintf("  Credenciales: %d/%d configuradas\n", configuredSecrets, len(m.secretFields)))
		b.WriteString("\n")
	}

	// Steps count
	b.WriteString(fmt.Sprintf("  Pasos de instalación: %d\n", len(m.resolvedSteps)))
	b.WriteString("\n")

	b.WriteString(boxStyle.Render("Se creará un backup antes de modificar archivos"))
	b.WriteString("\n\n")
	b.WriteString(normalStyle.Render("  Enter para instalar • Esc atrás • q cancelar"))
	return b.String()
}

func (m tuiModel) viewInstalling() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Instalando..."))
	b.WriteString("\n\n")

	// Show resolved steps with status
	for i, s := range m.resolvedSteps {
		var icon, name string
		if i < m.currentStep {
			icon = stepOkStyle.Render("✓")
			name = dimStyle.Render(s.Name)
		} else if i == m.currentStep {
			icon = stepActiveStyle.Render("◉")
			name = stepActiveStyle.Render(s.Name)
		} else {
			icon = stepPendingStyle.Render("○")
			name = stepPendingStyle.Render(s.Name)
		}
		b.WriteString(fmt.Sprintf("    %s %s\n", icon, name))
	}

	b.WriteString("\n")
	// Show last few logs
	start := 0
	if len(m.stepLogs) > 5 {
		start = len(m.stepLogs) - 5
	}
	for _, log := range m.stepLogs[start:] {
		b.WriteString(fmt.Sprintf("  %s\n", dimStyle.Render(log)))
	}

	return b.String()
}

func (m tuiModel) viewDone() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(successStyle.Render("  ✓ Instalación completada exitosamente"))
	b.WriteString("\n\n")

	if m.result != nil {
		b.WriteString(fmt.Sprintf("  %d pasos completados en %s\n\n", len(m.result.Completed), m.result.Duration.Round(1e6)))
	}

	b.WriteString(normalStyle.Render("  Próximos pasos:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("    1. %s\n", normalStyle.Render("Reinicia Claude Code para cargar los MCP servers")))
	b.WriteString(fmt.Sprintf("    2. %s\n", normalStyle.Render("Ejecuta 'inteliside verify' para confirmar")))
	b.WriteString(fmt.Sprintf("    3. %s\n", normalStyle.Render("Instala los plugins: /plugin marketplace add Intelliaa/marketplace-plugins-inteliside")))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Enter para salir"))
	return b.String()
}

func (m tuiModel) viewError() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(errorStyle.Render("  ✗ Error durante la instalación"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(fmt.Sprintf("  %s\n", normalStyle.Render(m.err.Error())))
	}

	if m.result != nil && len(m.result.Completed) > 0 {
		b.WriteString(fmt.Sprintf("\n  Pasos completados: %d\n", len(m.result.Completed)))
		b.WriteString("  Usa 'inteliside backup restore' para revertir cambios\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Enter para salir"))
	return b.String()
}

// helpers

func roleIcon(presetID string) string {
	switch presetID {
	case "pm":
		return "📋"
	case "designer":
		return "🎨"
	case "dev":
		return "⚡"
	case "fullstack":
		return "🚀"
	case "automation":
		return "🔄"
	case "legacy":
		return "🔧"
	case "custom":
		return "⚙️"
	default:
		return "•"
	}
}
