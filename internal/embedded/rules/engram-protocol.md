# Protocolo de memoria Engram (ATL Inteliside)

ATL Inteliside usa Engram con **dos proyectos separados** para aislar el conocimiento permanente del equipo del estado efímero del pipeline:

## Proyecto de equipo (`engram_project`)

Conocimiento permanente que sobrevive entre features y es consultado por todos los devs:

- Decisiones de arquitectura (`team/decisions/{tema}`)
- Patrones de código establecidos (`team/patterns/{area}`)
- Bugs resueltos con root cause (`team/bugs/{descripcion}`)
- Features completadas (`team/completed/{feature}`)

## Proyecto de pipeline (`engram_pipeline = "{engram_project}/atl"`)

Estado efímero del pipeline de implementación de la feature activa. Se genera automáticamente.

- Feature spec del PM, exploración, propuesta
- Spec técnica, design, tasks breakdown
- Estado del DAG (`pipeline-state`)
- Verify report

## Protocolo al arrancar sesión

```
1. mem_context(project: "{engram_project}/atl")
   → ¿Hay un pipeline activo?
2. mem_context(project: "{engram_project}")
   → Conocimiento del equipo
3. Si hay pipeline activo, recuperar pipeline-state y continuar
```

## Protocolo al terminar trabajo significativo

Claude guarda automáticamente en el proyecto correcto:
- Artefactos del pipeline → `{engram_project}/atl`
- Decisiones, patrones, bugs → `{engram_project}`
