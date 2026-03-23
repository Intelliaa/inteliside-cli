# Monitoreo de ventana de contexto (ATL Inteliside)

## Statusline

Cuando trabajes con ATL Inteliside, la statusline debe mostrar:

```
modelo | contexto usado% | fase actual del DAG
```

## Registro de contexto en pipeline-state

Al actualizar `pipeline-state` en Engram entre fases del DAG, incluir siempre el campo `context_pct` con el porcentaje aproximado de contexto usado en ese momento:

```
pipeline-state:
  feature: feat-{nombre}
  current_phase: sdd-apply
  apply_stage: Core
  context_pct: 42%
  phases_completed: [from-github, init, explore, propose, spec, design, tasks, write-tests]
```

Esto permite comparar consumo de contexto entre proyectos y detectar features que consumen más de lo esperado.

## Alerta al 75%

Si el contexto del orquestador supera el **75%**, informar al usuario:

> "El contexto del orquestador está al {N}%. Opciones:
> (a) Continuar — auto-compaction se activará si es necesario
> (b) Guardar estado y reanudar en sesión limpia"

**No forzar el corte** — dejar que el usuario decida. El pipeline-state en Engram ya contiene todo lo necesario para reanudar.

### Por qué 75% y no antes

- Los subagentes corren en `context: fork`, el orquestador se mantiene lean
- El buffer de autocompact es ~16.5%, hay margen hasta ~91.5%
- El benchmark de LinkVault (feature CRUD completa) consumió solo 35%
- Alertar antes del 75% generaría falsas alarmas en la mayoría de features
