package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
)

// SelectSkillsInteractively muestra la lista de skills y permite al usuario
// elegir por índices o "all". Lee desde r (normalmente os.Stdin).
func SelectSkillsInteractively(skills []model.SkillRef, r io.Reader) ([]model.SkillRef, error) {
	if len(skills) == 0 {
		return nil, nil
	}

	fmt.Fprintln(os.Stdout, "Skills disponibles:")
	for i, skill := range skills {
		fmt.Fprintf(os.Stdout, "%d. [%s] %s (%s)\n", i+1, skill.SourceID, skill.Name, skill.RepoPath)
	}
	fmt.Fprintln(os.Stdout, "Selecciona índices separados por coma (ej. 1,3,7) o 'all':")

	reader := bufio.NewReader(r)
	raw, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer la selección interactiva: %w", err)
	}
	choice := strings.ToLower(strings.TrimSpace(raw))
	if choice == "all" {
		return skills, nil
	}

	indexes := strings.Split(choice, ",")
	selected := make([]model.SkillRef, 0, len(indexes))
	seen := map[int]struct{}{}
	for _, index := range indexes {
		index = strings.TrimSpace(index)
		if index == "" {
			continue
		}
		n, err := strconv.Atoi(index)
		if err != nil {
			return nil, fmt.Errorf("índice inválido %q", index)
		}
		if n < 1 || n > len(skills) {
			return nil, fmt.Errorf("índice fuera de rango: %d", n)
		}
		if _, exists := seen[n]; exists {
			continue
		}
		seen[n] = struct{}{}
		selected = append(selected, skills[n-1])
	}
	return selected, nil
}
