# CLAUDE.md — {{project_name}}

> Configuracion del proyecto para Claude Code y ATL Inteliside.
> Completa los valores marcados con `<!-- TODO: -->`.

---

<!-- inteliside:atl-config -->
## ATL Inteliside

Configuracion requerida para el plugin ATL Inteliside. Todos los devs del equipo deben
tener este archivo con los mismos valores para compartir la memoria de Engram.

```yaml
engram_project: "{{engram_project}}"
github_owner: "{{github_owner}}"
github_repo: "{{github_repo}}"
```

> **Nota**: ATL Inteliside deriva automaticamente un segundo proyecto Engram para el pipeline:
> `engram_pipeline = "{engram_project}/atl"`
>
> - **`engram_project`** → conocimiento permanente del equipo (decisiones, patrones, bugs)
> - **`engram_pipeline`** → estado efimero del pipeline de implementacion (specs, design, tasks, verify-report)
>
> No necesitas configurar el pipeline — el orquestador lo genera automaticamente.
<!-- /inteliside:atl-config -->

---

## Proyecto

- **Nombre**: {{project_name}}
- **Descripcion**: {{project_description}}
- **Stack**: {{project_stack}}
- **Entorno de desarrollo**: {{dev_environment}}

---

## Comandos frecuentes

```bash
# Desarrollo
{{cmd_dev}}

# Tests
{{cmd_test}}

# Build
{{cmd_build}}

# DB (si aplica)
{{cmd_db}}
```

---

## Convenciones del proyecto

### Estructura de carpetas
```
{{folder_structure}}
```

### Naming
- Archivos: `kebab-case.ts`
- Componentes: `PascalCase.tsx`
- Funciones: `camelCase`
- Variables de entorno: `UPPER_SNAKE_CASE`

### Testing
- Framework: {{test_framework}}
- Archivos de test: `*.test.ts` junto al archivo testeado
- Coverage minimo: 80%

---

## Variables de entorno requeridas

Ver `.env.example` para la lista completa. Las criticas son:

```bash
{{env_vars}}
```

---

## Rules de ATL Inteliside

- @.claude/rules/engram-protocol.md
- @.claude/rules/subagent-architecture.md
- @.claude/rules/atl-workflow.md
- @.claude/rules/context-monitoring.md
- @.claude/rules/team-rules.md

---

*Generado con inteliside init — Marketplace Inteliside*
