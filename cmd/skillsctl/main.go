package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/config"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/filter"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/install"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/source"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/state"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/ui"
	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "init":
		err = runInit(os.Args[2:])
	case "add-source":
		err = runAddSource(os.Args[2:])
	case "list-sources":
		err = runListSources(os.Args[2:])
	case "sync":
		err = runSync(os.Args[2:])
	case "browse":
		err = runBrowse(os.Args[2:])
	case "select":
		err = runSelect(os.Args[2:])
	case "install":
		err = runInstall(os.Args[2:])
	case "update":
		err = runUpdate(os.Args[2:])
	case "remove":
		err = runRemove(os.Args[2:])
	case "inspect":
		err = runInspect(os.Args[2:])
	case "export":
		err = runExport(os.Args[2:])
	case "-h", "--help", "help":
		printHelp()
		return
	default:
		err = fmt.Errorf("comando desconocido: %s", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultConfigPath, "ruta del archivo de configuración")
	fromLinks := fs.String("from-links", "", "archivo markdown/txt con enlaces de repositorios")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if _, err := os.Stat(*configPath); err == nil {
		return fmt.Errorf("el archivo de configuración ya existe: %s", *configPath)
	}

	cfg := config.DefaultConfig()
	if strings.TrimSpace(*fromLinks) != "" {
		imported, err := config.BuildConfigFromLinksFile(*fromLinks)
		if err != nil {
			return err
		}
		cfg = imported
	}
	if err := config.Save(*configPath, cfg); err != nil {
		return err
	}
	fmt.Printf("Configuración creada en %s\n", *configPath)
	return nil
}

func runAddSource(args []string) error {
	fs := flag.NewFlagSet("add-source", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultConfigPath, "ruta del archivo de configuración")
	id := fs.String("id", "", "id único de la fuente")
	url := fs.String("url", "", "url del repositorio fuente")
	defRef := fs.String("ref", "", "rama o tag por defecto")
	tags := fs.String("tags", "", "tags por defecto separados por coma")
	roles := fs.String("roles", "", "roles por defecto separados por coma")
	useCases := fs.String("use-cases", "", "casos de uso por defecto separados por coma")
	areas := fs.String("areas", "", "áreas temáticas por defecto separadas por coma")
	priority := fs.Int("priority", 0, "prioridad por defecto")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := safeLoadConfig(*configPath)
	if err != nil {
		return err
	}
	src := config.SourceConfig{
		ID:           strings.TrimSpace(*id),
		URL:          strings.TrimSpace(*url),
		Enabled:      true,
		DefaultRef:   strings.TrimSpace(*defRef),
		DefaultTags:  parseCSV(*tags),
		DefaultRoles: parseCSV(*roles),
		DefaultCases: parseCSV(*useCases),
		DefaultAreas: parseCSV(*areas),
		DefaultPrio:  *priority,
	}
	if err := cfg.AddSource(src); err != nil {
		return err
	}
	if err := config.Save(*configPath, cfg); err != nil {
		return err
	}
	fmt.Printf("Fuente %q agregada\n", src.ID)
	return nil
}

func runListSources(args []string) error {
	fs := flag.NewFlagSet("list-sources", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultConfigPath, "ruta del archivo de configuración")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if len(cfg.Sources) == 0 {
		fmt.Println("No hay fuentes configuradas.")
		return nil
	}
	for i, src := range cfg.Sources {
		status := "disabled"
		if src.Enabled {
			status = "enabled"
		}
		fmt.Printf("%d. %s [%s] -> %s\n", i+1, src.ID, status, src.URL)
	}
	return nil
}

func runSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultConfigPath, "ruta del archivo de configuración")
	sources := fs.String("sources", "", "ids de fuente separados por coma")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	skills, warnings, err := discoverAll(cfg, parseCSV(*sources))
	if err != nil {
		return err
	}
	path, err := state.SaveCatalog(skills)
	if err != nil {
		return err
	}
	for _, warn := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warn)
	}
	fmt.Printf("Catálogo sincronizado: %d skills (%s)\n", len(skills), path)
	return nil
}

func runBrowse(args []string) error {
	fs := flag.NewFlagSet("browse", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultConfigPath, "ruta del archivo de configuración")
	useCache := fs.Bool("use-cache", false, "usar catálogo local sincronizado")
	sources := fs.String("sources", "", "ids de fuente separados por coma")
	tags := fs.String("tags", "", "filtro por tags")
	roles := fs.String("roles", "", "filtro por roles")
	useCases := fs.String("use-cases", "", "filtro por casos de uso")
	areas := fs.String("areas", "", "filtro por áreas")
	projects := fs.String("projects", "", "filtro por proyectos")
	minPriority := fs.Int("min-priority", 0, "prioridad mínima")
	maxSkills := fs.Int("max", 0, "máximo de skills en salida")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var skills []model.SkillRef
	var warnings []string
	if *useCache {
		catalog, _, err := state.LoadCatalog()
		if err != nil {
			return err
		}
		skills = catalog.Skills
	} else {
		cfg, err := config.Load(*configPath)
		if err != nil {
			return err
		}
		skills, warnings, err = discoverAll(cfg, parseCSV(*sources))
		if err != nil {
			return err
		}
	}

	filtered := filter.Apply(skills, filter.Query{
		Sources:     parseCSV(*sources),
		Tags:        parseCSV(*tags),
		Roles:       parseCSV(*roles),
		UseCases:    parseCSV(*useCases),
		Areas:       parseCSV(*areas),
		Projects:    parseCSV(*projects),
		MinPriority: *minPriority,
		MaxSkills:   *maxSkills,
	})

	for _, warn := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warn)
	}
	if len(filtered) == 0 {
		fmt.Println("No se encontraron skills con los filtros actuales.")
		return nil
	}
	for i, skill := range filtered {
		fmt.Printf("%d. [%s] %s | path=%s | tags=%s | roles=%s | prio=%d\n",
			i+1, skill.SourceID, skill.Name, skill.RepoPath, strings.Join(skill.Tags, ","), strings.Join(skill.Roles, ","), skill.Priority)
	}
	return nil
}

func runSelect(args []string) error {
	fs := flag.NewFlagSet("select", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultConfigPath, "ruta del archivo de configuración")
	project := fs.String("project", "", "nombre de proyecto")
	useCache := fs.Bool("use-cache", true, "usar catálogo local sincronizado")
	ids := fs.String("ids", "", "ids de skills separados por coma (modo no interactivo)")
	sources := fs.String("sources", "", "ids de fuente")
	tags := fs.String("tags", "", "filtro por tags")
	roles := fs.String("roles", "", "filtro por roles")
	useCases := fs.String("use-cases", "", "filtro por casos de uso")
	areas := fs.String("areas", "", "filtro por áreas")
	minPriority := fs.Int("min-priority", 0, "prioridad mínima")
	maxSkills := fs.Int("max", 0, "límite de skills")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*project) == "" {
		return errors.New("debes indicar --project")
	}

	var skills []model.SkillRef
	var err error
	if *useCache {
		catalog, _, err := state.LoadCatalog()
		if err != nil {
			return err
		}
		skills = catalog.Skills
	} else {
		cfg, err := config.Load(*configPath)
		if err != nil {
			return err
		}
		skills, _, err = discoverAll(cfg, parseCSV(*sources))
		if err != nil {
			return err
		}
	}

	filtered := filter.Apply(skills, filter.Query{
		Sources:     parseCSV(*sources),
		Tags:        parseCSV(*tags),
		Roles:       parseCSV(*roles),
		UseCases:    parseCSV(*useCases),
		Areas:       parseCSV(*areas),
		MinPriority: *minPriority,
		MaxSkills:   *maxSkills,
	})
	if len(filtered) == 0 {
		return errors.New("no hay skills candidatas para seleccionar")
	}

	var selected []model.SkillRef
	if explicitIDs := parseCSV(*ids); len(explicitIDs) > 0 {
		selected = selectByIDs(filtered, explicitIDs)
		if len(selected) == 0 {
			return errors.New("ningún id seleccionado coincide con skills filtradas")
		}
	} else {
		selected, err = ui.SelectSkillsInteractively(filtered, os.Stdin)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			return errors.New("no se seleccionaron skills")
		}
	}

	path, err := state.SaveSelection(*project, selected)
	if err != nil {
		return err
	}
	fmt.Printf("Selección guardada para proyecto %q en %s (%d skills)\n", *project, path, len(selected))
	return nil
}

func runInstall(args []string) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	project := fs.String("project", "", "nombre de proyecto")
	selectionFile := fs.String("selection", "", "ruta de archivo de selección")
	outputDir := fs.String("output", ".skills", "directorio local de instalación")
	layout := fs.String("layout", "default", "layout local: default o claude")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var (
		selection model.SelectionFile
		err       error
	)
	switch {
	case strings.TrimSpace(*selectionFile) != "":
		selection, err = state.LoadSelection(*selectionFile)
	case strings.TrimSpace(*project) != "":
		selection, err = state.LoadSelection(filepath.Join(state.SelectionDir, *project+".yaml"))
	default:
		return errors.New("debes indicar --project o --selection")
	}
	if err != nil {
		return err
	}

	installer := install.NewInstaller()
	result := installer.InstallSkills(selection.Skills, *outputDir, install.Layout(*layout))
	if len(result.Records) == 0 {
		return fmt.Errorf("no se instaló ninguna skill; errores=%d", len(result.Errors))
	}

	installState := model.InstallState{
		Version: "1",
		Project: selection.Project,
		Records: result.Records,
	}
	statePath, err := state.SaveInstallState(selection.Project, installState)
	if err != nil {
		return err
	}
	for _, installErr := range result.Errors {
		fmt.Fprintf(os.Stderr, "warning: %v\n", installErr)
	}
	fmt.Printf("Instalación completada: %d skills (%s)\n", len(result.Records), statePath)
	return nil
}

func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	project := fs.String("project", "", "nombre de proyecto")
	layout := fs.String("layout", "default", "layout local: default o claude")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*project) == "" {
		return errors.New("debes indicar --project")
	}

	currentState, _, err := state.LoadInstallState(*project)
	if err != nil {
		return err
	}
	if len(currentState.Records) == 0 {
		return errors.New("no hay skills instaladas para actualizar")
	}

	skills := make([]model.SkillRef, 0, len(currentState.Records))
	for _, rec := range currentState.Records {
		skills = append(skills, model.SkillRef{
			ID:       rec.SkillID,
			SourceID: rec.SourceID,
			RepoURL:  rec.RepoURL,
			RepoPath: rec.RepoPath,
			Ref:      rec.Ref,
			Name:     filepath.Base(rec.RepoPath),
		})
	}
	installer := install.NewInstaller()
	result := installer.InstallSkills(skills, ".skills", install.Layout(*layout))
	if len(result.Records) == 0 {
		return fmt.Errorf("no se pudo actualizar ninguna skill; errores=%d", len(result.Errors))
	}
	newState := model.InstallState{
		Version: "1",
		Project: *project,
		Records: result.Records,
	}
	statePath, err := state.SaveInstallState(*project, newState)
	if err != nil {
		return err
	}
	for _, updateErr := range result.Errors {
		fmt.Fprintf(os.Stderr, "warning: %v\n", updateErr)
	}
	fmt.Printf("Actualización completada: %d skills (%s)\n", len(result.Records), statePath)
	return nil
}

func runRemove(args []string) error {
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultConfigPath, "ruta del archivo de configuración")
	sourceID := fs.String("source-id", "", "id de fuente a eliminar")
	project := fs.String("project", "", "proyecto para eliminar selección/estado")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*sourceID) == "" && strings.TrimSpace(*project) == "" {
		return errors.New("debes indicar --source-id o --project")
	}

	if strings.TrimSpace(*sourceID) != "" {
		cfg, err := config.Load(*configPath)
		if err != nil {
			return err
		}
		filtered := make([]config.SourceConfig, 0, len(cfg.Sources))
		removed := false
		for _, src := range cfg.Sources {
			if src.ID == *sourceID {
				removed = true
				continue
			}
			filtered = append(filtered, src)
		}
		if !removed {
			return fmt.Errorf("no existe una fuente con id %q", *sourceID)
		}
		cfg.Sources = filtered
		if err := config.Save(*configPath, cfg); err != nil {
			return err
		}
		fmt.Printf("Fuente %q eliminada\n", *sourceID)
	}

	if strings.TrimSpace(*project) != "" {
		selectionPath := filepath.Join(state.SelectionDir, *project+".yaml")
		statePath := filepath.Join(state.StateDir, *project+".json")
		_ = os.Remove(selectionPath)
		_ = os.Remove(statePath)
		fmt.Printf("Metadatos del proyecto %q eliminados\n", *project)
	}
	return nil
}

func runInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	id := fs.String("id", "", "id de skill")
	project := fs.String("project", "", "proyecto para inspeccionar instalación")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*id) == "" && strings.TrimSpace(*project) == "" {
		return errors.New("usa --id para una skill o --project para estado de instalación")
	}

	if strings.TrimSpace(*id) != "" {
		catalog, _, err := state.LoadCatalog()
		if err != nil {
			return err
		}
		for _, skill := range catalog.Skills {
			if skill.ID != *id {
				continue
			}
			content, _ := yaml.Marshal(skill)
			fmt.Print(string(content))
			return nil
		}
		return fmt.Errorf("skill %q no encontrada en catálogo local", *id)
	}

	installState, _, err := state.LoadInstallState(*project)
	if err != nil {
		return err
	}
	content, _ := yaml.Marshal(installState)
	fmt.Print(string(content))
	return nil
}

func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	project := fs.String("project", "", "proyecto")
	selectionFile := fs.String("selection", "", "archivo de selección")
	output := fs.String("output", "", "ruta de exportación yaml")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*project) == "" {
		return errors.New("debes indicar --project")
	}

	if strings.TrimSpace(*selectionFile) == "" {
		*selectionFile = filepath.Join(state.SelectionDir, *project+".yaml")
	}
	selection, err := state.LoadSelection(*selectionFile)
	if err != nil {
		return err
	}
	profile := config.ProjectProfile{
		Name:      *project,
		Sources:   uniqueSources(selection.Skills),
		MaxSkills: len(selection.Skills),
	}
	payload, err := yaml.Marshal(profile)
	if err != nil {
		return err
	}

	if strings.TrimSpace(*output) == "" {
		fmt.Print(string(payload))
		return nil
	}
	if err := os.WriteFile(*output, payload, 0o644); err != nil {
		return fmt.Errorf("no se pudo exportar perfil: %w", err)
	}
	fmt.Printf("Perfil exportado en %s\n", *output)
	return nil
}

func discoverAll(cfg config.Config, sourceIDs []string) ([]model.SkillRef, []string, error) {
	enabled := cfg.EnabledSources(sourceIDs)
	if len(enabled) == 0 {
		return nil, nil, errors.New("no hay fuentes habilitadas para procesar")
	}
	discovery := source.NewDiscovery()
	allSkills := []model.SkillRef{}
	seenIDs := map[string]struct{}{}
	warnings := []string{}

	for _, src := range enabled {
		skills, err := discovery.Discover(src)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("fuente %s omitida: %v", src.ID, err))
			continue
		}
		for _, skill := range skills {
			if _, exists := seenIDs[skill.ID]; exists {
				warnings = append(warnings, fmt.Sprintf("skill duplicada descartada: %s", skill.ID))
				continue
			}
			seenIDs[skill.ID] = struct{}{}
			allSkills = append(allSkills, skill)
		}
	}

	if len(allSkills) == 0 {
		return nil, warnings, errors.New("no se pudo descubrir ninguna skill utilizable")
	}
	sort.Slice(allSkills, func(i, j int) bool { return allSkills[i].ID < allSkills[j].ID })
	return allSkills, warnings, nil
}

func safeLoadConfig(path string) (config.Config, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return config.DefaultConfig(), nil
	}
	cfg, err := config.Load(path)
	if err == nil {
		return cfg, nil
	}
	return config.Config{}, err
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		normalized := strings.ToLower(strings.TrimSpace(part))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func selectByIDs(skills []model.SkillRef, ids []string) []model.SkillRef {
	selected := []model.SkillRef{}
	idSet := map[string]struct{}{}
	for _, id := range ids {
		idSet[strings.TrimSpace(id)] = struct{}{}
	}
	for _, skill := range skills {
		if _, ok := idSet[skill.ID]; ok {
			selected = append(selected, skill)
		}
	}
	return selected
}

func uniqueSources(skills []model.SkillRef) []string {
	set := map[string]struct{}{}
	for _, skill := range skills {
		set[skill.SourceID] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func printHelp() {
	fmt.Printf(`skillsctl - gestor modular de skills remotas

Comandos:
  init         Crea configuración base
  add-source   Agrega fuente remota
  list-sources Lista fuentes registradas
  sync         Descubre skills y guarda catálogo local mínimo
  browse       Explora skills con filtros
  select       Selecciona skills (interactivo o por IDs)
  install      Descarga solo skills seleccionadas bajo demanda
  update       Actualiza skills instaladas desde origen remoto
  remove       Elimina fuente o metadatos de proyecto
  inspect      Inspecciona una skill o estado de instalación
  export       Exporta perfil YAML de proyecto

Ejemplo rápido:
  skillsctl init --from-links sources.md
  skillsctl sync
  skillsctl browse --tags go,cli
  skillsctl select --project myproj
  skillsctl install --project myproj --layout claude

Timestamp: %s
`, time.Now().Format(time.RFC3339))
}
