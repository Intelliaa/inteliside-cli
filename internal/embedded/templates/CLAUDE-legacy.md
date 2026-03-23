# Contexto Legacy — {{project_name}}

> Este directorio contiene los artefactos del onboarding legacy.
> Generados por los wizards de SDD-Legacy antes de migrar a ATL Inteliside.
>
> **No eliminar** — estos archivos son referencia para la implementacion de nuevas features.

---

## Origen

Este proyecto fue onboarded desde un codebase legacy via SDD-Legacy.
Los artefactos aqui documentados fueron generados durante el proceso de auditoria,
extraccion de reglas de negocio y formalizacion de convenciones.

---

## Artefactos esperados

| Archivo | Descripcion | Generado por |
|---------|-------------|--------------|
| `CLAUDE.md` | Configuracion original del proyecto | legacy-onboard-wizard |
| `.claude/rules/` | Rules originales formalizadas | legacy-conventions-wizard |
| `audit-report.md` | Health score y deuda tecnica | legacy-audit-wizard |
| `business-rules/` | Reglas de negocio extraidas del codigo | legacy-rules-extractor |
| `conventions.md` | Convenciones formalizadas del codebase | legacy-conventions-wizard |
| `feature-specs/` | Features reverse-engineered | legacy-spec-extractor |
| `baseline-plan.md` | Plan de accion priorizado | legacy-baseline-wizard |

---

## Como usar esta documentacion

1. **Antes de implementar una feature nueva**: consultar `business-rules/` para reglas de negocio existentes
2. **Antes de refactorizar**: consultar `CLAUDE.md` original para convenciones del codebase
3. **Para contexto historico**: el `audit-report.md` documenta el estado del proyecto al momento del onboarding

---

## Archivo CLAUDE.md original

El CLAUDE.md original del proyecto (pre-ATL) fue archivado aqui.
La configuracion activa del proyecto ahora vive en el CLAUDE.md raiz.

---

*Archivado por inteliside init — Marketplace Inteliside*
