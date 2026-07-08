# skillsctl (CLI modular de skills remotas)

`skillsctl` es una herramienta en Go para descubrir, seleccionar e instalar skills desde repositorios remotos bajo demanda, evitando clonar y guardar todo el contenido remoto por defecto.

## 1) Arquitectura y contrato de comportamiento

### Arquitectura modular

1. **CLI (`cmd/skillsctl`)**
   1. Parsea comandos/flags.
   2. Orquesta flujo de configuración, descubrimiento, filtros, selección e instalación.
2. **Configuración (`internal/config`)**
   1. Carga/valida YAML.
   2. Importa enlaces desde Markdown/txt para construir configuración inicial.
3. **Resolución de fuentes (`internal/source`)**
   1. Soporta fuentes GitHub.
   2. Descubre skills por análisis de árbol remoto (API GitHub, sin clone completo).
4. **Filtrado (`internal/filter`)**
   1. Filtra por `sources`, `tags`, `roles`, `use_cases`, `areas`, `projects`, `priority`, `max`.
5. **Selección interactiva (`internal/ui`)**
   1. UI de terminal con listas numeradas y selección por índices.
6. **Instalación bajo demanda (`internal/install`)**
   1. Descarga solo directorios/archivos de skills seleccionadas.
   2. Soporta layout `default` y `claude`.
7. **Estado mínimo/auditoría (`internal/state`)**
   1. Catálogo local mínimo (`sync`).
   2. Selecciones por proyecto.
   3. Estado de instalación para `update` e inspección.
8. **Modelos (`internal/model`)**
   1. Contratos de datos compartidos.

### Contrato de comportamiento

1. **No descarga masiva por defecto**: solo metadatos en `sync` y contenido en `install/update`.
2. **Descarga bajo demanda**: una skill se descarga al instalarse o actualizarse.
3. **Manejo de errores parcial**: si una fuente falla, se reporta warning y se continúa con otras.
4. **Detección de duplicados**: IDs duplicados se descartan con warning.
5. **Auditable y repetible**: cada selección e instalación genera archivos de estado mínimos.
6. **Extensible**: nuevas fuentes (GitLab, local, etc.) se pueden añadir en `internal/source`.

## 2) Estructura de carpetas

```text
.
├─ cmd/
│  └─ skillsctl/
│     └─ main.go
├─ internal/
│  ├─ config/
│  │  ├─ config.go
│  │  └─ importer.go
│  ├─ filter/
│  │  └─ filter.go
│  ├─ install/
│  │  └─ installer.go
│  ├─ model/
│  │  └─ types.go
│  ├─ source/
│  │  ├─ discovery.go
│  │  └─ github.go
│  ├─ state/
│  │  ├─ catalog.go
│  │  └─ files.go
│  └─ ui/
│     └─ select.go
├─ go.mod
└─ README.md
```

## 3) Diseño de comandos CLI

Comandos implementados:

1. `init` crea configuración base (`.skillsctl\config.yaml`) e importa enlaces desde markdown/txt.
2. `add-source` registra una fuente con defaults de taxonomía.
3. `list-sources` lista fuentes registradas.
4. `sync` descubre skills remotas y guarda catálogo local mínimo.
5. `browse` explora skills con filtros.
6. `select` permite selección interactiva o por IDs y guarda selección por proyecto.
7. `install` descarga solo las skills seleccionadas y prepara estructura local.
8. `update` actualiza skills instaladas desde su origen remoto.
9. `remove` elimina una fuente o metadatos de proyecto.
10. `inspect` muestra detalle de skill o estado de instalación.
11. `export` exporta perfil YAML de proyecto.

## 4) Formato de configuración recomendado

Archivo: `.skillsctl\config.yaml`

```yaml
version: "1"
sources:
  - id: miguel-code-skills
    url: https://github.com/MIGUEL-MOTTA-U/code-skills
    enabled: true
    default_ref: main
    default_tags: [go, cli]
    default_roles: [backend]
    default_use_cases: [dev-env]
    default_areas: [automation]
    default_priority: 5
    projects: [customer-a, customer-b]
    classifiers:
      - path_prefix: skills/security
        tags: [security]
        roles: [security-engineer]
        use_cases: [audit]
        areas: [cybersecurity]
        priority: 9
        projects: [customer-a]
projects:
  - name: customer-a
    sources: [miguel-code-skills]
    tags: [security]
    roles: [security-engineer]
    use_cases: [audit]
    max_skills: 10
    min_priority: 5
    output_directory: .skills
```

## 5) Flujo completo de uso

1. `skillsctl init --from-links sources.md`
2. `skillsctl add-source --id openai-skills --url https://github.com/openai/skills`
3. `skillsctl list-sources`
4. `skillsctl sync`
5. `skillsctl browse --tags go,cli --roles backend --max 20`
6. `skillsctl select --project customer-a`
7. `skillsctl install --project customer-a --layout claude`
8. `skillsctl update --project customer-a`
9. `skillsctl inspect --project customer-a`
10. `skillsctl export --project customer-a --output customer-a-profile.yaml`

## 6) Implementación inicial funcional

La implementación actual cubre:

1. Fuentes múltiples en YAML.
2. Importación automática desde markdown/txt de enlaces GitHub.
3. Clasificación por roles, tags, casos de uso, áreas, prioridad y proyectos (defaults + reglas por prefijo de path).
4. Exploración y selección interactiva por terminal.
5. Filtros por flags (`sources`, `tags`, `roles`, `use-cases`, `areas`, `min-priority`, `max`).
6. Descarga bajo demanda únicamente de skills seleccionadas.
7. Preparación de estructura local compatible (`default` o `.claude\skills`).
8. Estado mínimo para repetir, auditar y actualizar.

## 7) Manejo de errores y validaciones

Se manejan explícitamente:

1. URLs inválidas o no soportadas.
2. Repositorios inaccesibles o respuesta remota inválida.
3. Fuentes sin estructura compatible descubierta.
4. Duplicación de IDs de fuente o skills.
5. Conflictos y errores parciales por fuente (se continúa con otras).
6. Errores de descarga por archivo/directorio.
7. Inputs inválidos en selección interactiva y flags.

## 8) Supuestos y justificación de decisiones

Supuestos declarados (para no inventar estructura de repos):

1. **Descubrimiento heurístico**: se detectan skills en rutas que contengan `skill`/`skills` y archivos como `skill.yaml|yml|json|md` o `README.md`.
2. **Fuente inicial soportada**: GitHub público vía API (extensible a nuevas fuentes).
3. **Clasificación principal**: se prioriza taxonomía configurable en `config.yaml`, porque los repos pueden tener estructuras heterogéneas.

Justificación:

1. **Simplicidad y mantenimiento**: paquetes pequeños y responsabilidades separadas.
2. **Eficiencia**: no hay clone completo ni mirror local permanente.
3. **Extensibilidad**: contrato claro para añadir resolutores y clasificadores más ricos en siguientes iteraciones.

## 9) Compilación e Instalación Multiplataforma (Linux, Windows 11, Mac)

Para integrar y reutilizar la herramienta `skillsctl` en cualquier sistema operativo, puedes compilar el binario localmente o mediante compilación cruzada.

### A. Pasos para Compilar

Asegúrate de tener Go instalado (versión 1.22 o superior). Ejecuta el comando correspondiente desde la raíz del repositorio:

#### 1. Windows 11 (PowerShell / CMD)
```powershell
go build -o bin/skillsctl.exe ./cmd/skillsctl
```

#### 2. Linux (Bash / Zsh)
```bash
go build -o bin/skillsctl ./cmd/skillsctl
```

#### 3. macOS (Apple Silicon M1/M2/M3 o Intel)
```bash
go build -o bin/skillsctl ./cmd/skillsctl
```

### B. Compilación Cruzada (Cross-Compilation)
Si estás en una plataforma (ej. Windows) y deseas generar los binarios para todas las demás:

```powershell
# Para Linux (64-bit)
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/skillsctl-linux ./cmd/skillsctl

# Para macOS (Apple Silicon - ARM64)
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/skillsctl-darwin-arm64 ./cmd/skillsctl

# Para macOS (Intel - AMD64)
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o bin/skillsctl-darwin-amd64 ./cmd/skillsctl

# Para Windows (64-bit)
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/skillsctl.exe ./cmd/skillsctl
```

### C. Integración en el PATH del Sistema

Para poder ejecutar `skillsctl` desde cualquier directorio del sistema sin escribir la ruta completa:

#### En Windows 11:
1. Crea un directorio permanente, por ejemplo: `C:\bin\`.
2. Mueve el archivo `skillsctl.exe` a ese directorio.
3. Abre el menú de inicio, busca **"Variables de entorno del sistema"** y selecciónalo.
4. Haz clic en **"Variables de entorno..."**.
5. En la sección "Variables del usuario" o "Variables del sistema", selecciona la variable **Path** y haz clic en **Editar...**.
6. Haz clic en **Nuevo** y añade la ruta: `C:\bin`.
7. Haz clic en **Aceptar** en todas las ventanas abiertas. Reinicia tu terminal (PowerShell o CMD).

#### En Linux y macOS:
1. Mueve el binario compilado a una ruta en tu PATH (como `/usr/local/bin` o `~/bin`):
   ```bash
   sudo mv bin/skillsctl /usr/local/bin/
   sudo chmod +x /usr/local/bin/skillsctl
   ```
2. Si prefieres usar un directorio propio (ej. `~/bin`):
   ```bash
   mkdir -p ~/bin
   mv bin/skillsctl ~/bin/
   chmod +x ~/bin/skillsctl
   ```
   Asegúrate de agregar `export PATH="$HOME/bin:$PATH"` en tu archivo de configuración de terminal (`~/.bashrc`, `~/.zshrc` o `~/.bash_profile`).

---

## 10) Ejemplos Detallados de Configuración y Uso (16 Casos)

A continuación se presentan una serie de ejemplos para aprovechar al máximo todas las configuraciones y comandos que posee la herramienta.

### Ejemplo 1: Inicialización simple con configuración por defecto
Crea la estructura de directorio `.skillsctl` y el archivo `config.yaml` vacío con la versión inicial.
```bash
skillsctl init
```

### Ejemplo 2: Inicialización importando enlaces desde un archivo Markdown/TXT
Lee un archivo (como `sources.md`) que contiene enlaces de repositorios de GitHub (formato `https://github.com/owner/repo`) y genera un archivo de configuración base registrándolos automáticamente como fuentes habilitadas con identificadores normalizados.
```bash
skillsctl init --from-links sources.md
```

### Ejemplo 3: Registrar una fuente manualmente con taxonomía por defecto
Agrega una nueva fuente especificando una URL, ID, y clasificaciones predeterminadas para todas las skills que se descubran dentro del repositorio.
```bash
skillsctl add-source --id general-skills --url https://github.com/user/my-skills --ref main --tags go,cli --roles backend --priority 7
```

### Ejemplo 4: Listar fuentes registradas y su estado
Muestra todas las fuentes registradas en tu configuración junto a su estado (`enabled` o `disabled`) y sus URLs asociadas.
```bash
skillsctl list-sources
```

### Ejemplo 5: Sincronizar el catálogo local mínimo (Modo Completo o de Fuentes Específicas)
Escanea los repositorios habilitados para descubrir la estructura de carpetas de las skills. Guarda la información en `.skillsctl/state/catalog.json` sin descargar el contenido de los archivos de cada skill.
* **Sincronizar todo:**
  ```bash
  skillsctl sync
  ```
* **Sincronizar solo fuentes específicas (separadas por coma):**
  ```bash
  skillsctl sync --sources general-skills,miguel-code-skills
  ```

### Ejemplo 6: Explorar skills filtrando por taxonomía (browse sin caché)
Realiza descubrimiento remoto en tiempo real y muestra los resultados filtrando por tags o roles.
```bash
skillsctl browse --tags go,cli --roles backend
```

### Ejemplo 7: Explorar skills utilizando el catálogo local sincronizado (Caché rápida)
Busca y filtra skills de forma instantánea leyendo la caché sincronizada en lugar de consultar la API de GitHub en tiempo real.
```bash
skillsctl browse --use-cache --tags security --min-priority 5
```

### Ejemplo 8: Filtrado avanzado y límite de resultados
Muestra un máximo de 5 skills con prioridad superior a 3 del proyecto `customer-a`.
```bash
skillsctl browse --use-cache --min-priority 3 --max 5 --projects customer-a
```

### Ejemplo 9: Selección interactiva de skills para un proyecto
Permite seleccionar de manera interactiva por índices (ej. `1,2,3` o `all`) cuáles de las skills filtradas deseas asociar a un proyecto. Guarda la selección en `.skillsctl/selections/mi-proyecto.yaml`.
```bash
skillsctl select --project mi-proyecto --tags go --use-cache
```

### Ejemplo 10: Selección no interactiva por IDs explícitos (Automatización)
Permite realizar la selección de skills de manera directa especificando una lista de IDs separados por comas sin abrir la consola interactiva.
```bash
skillsctl select --project mi-proyecto --ids general-skills-skills-go-helper,miguel-code-skills-skills-security-audit
```

### Ejemplo 11: Instalar bajo demanda con layout por defecto (Default Layout)
Descarga el contenido completo de las carpetas de las skills seleccionadas para el proyecto. Estructura de carpetas resultante: `.skills/<source-id>/<skill-name-normalizado>/...`
```bash
skillsctl install --project mi-proyecto --output .skills --layout default
```

### Ejemplo 12: Instalar bajo demanda con layout optimizado para Claude (.claude Layout)
Instala las skills seleccionadas bajo el directorio especial `.claude/skills/<skill-name-normalizado>/...` lo cual permite que agentes como Claude las consuman de forma nativa.
```bash
skillsctl install --project mi-proyecto --layout claude
```

### Ejemplo 13: Actualizar skills instaladas desde su origen remoto
Actualiza el código de todas las skills que ya se encuentran instaladas para un proyecto trayendo los últimos cambios de sus respectivas ramas remotas.
```bash
skillsctl update --project mi-proyecto --layout claude
```

### Ejemplo 14: Inspección de metadatos de skill o estado de instalación del proyecto
* **Ver detalles de una skill específica del catálogo:**
  ```bash
  skillsctl inspect --id general-skills-skills-go-helper
  ```
* **Ver estado completo de las skills instaladas en un proyecto:**
  ```bash
  skillsctl inspect --project mi-proyecto
  ```

### Ejemplo 15: Exportar perfil del proyecto
Genera un perfil YAML resumen de las fuentes y número de skills utilizadas en el proyecto.
```bash
skillsctl export --project mi-proyecto --output mi-proyecto-resumen.yaml
```

### Ejemplo 16: Eliminar fuentes o metadatos de proyectos
* **Eliminar una fuente de la configuración:**
  ```bash
  skillsctl remove --source-id general-skills
  ```
* **Eliminar archivos de selección y estado asociados a un proyecto:**
  ```bash
  skillsctl remove --project mi-proyecto
  ```
