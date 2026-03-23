# Arquitectura de subagentes (ATL Inteliside)

ATL usa 5 subagentes especializados que corren en contexto aislado (`context: fork`).
El orquestador nunca ejecuta código — solo coordina y mantiene `pipeline-state`.

| Subagente | Rol | Skills |
|-----------|-----|--------|
| **atl-analyst** | Análisis y preparación | sdd-from-github, sdd-init, sdd-explore, sdd-propose |
| **atl-architect** | Especificación y planificación | sdd-spec, sdd-design, sdd-tasks |
| **atl-test-writer** | Tests independientes (antes de implementar) | sdd-write-tests |
| **atl-builder** | Implementación (TDD) | sdd-apply |
| **atl-verifier** | Verificación y cierre | sdd-verify, sdd-archive |
