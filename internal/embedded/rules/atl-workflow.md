# Flujo de trabajo con ATL Inteliside

El PM sube Feature Specs como milestones en GitHub Projects.
El equipo dev implementa con:

```
/atl-inteliside:orchestrador
→ /sdd-from-github [nombre-del-milestone]      → atl-analyst
→ /sdd-continue    (explore → propose)          → atl-analyst
→ /sdd-continue    (spec + design en paralelo)  → atl-architect
→ /sdd-continue    (tasks → crea issues)        → atl-architect
→ /sdd-write-tests                              → atl-test-writer
→ /sdd-apply Foundation                         → atl-builder
→ /sdd-apply Core                               → atl-builder
→ /sdd-apply Integration                        → atl-builder
→ /sdd-apply UI                                 → atl-builder
→ /sdd-apply Testing                            → atl-builder
→ /sdd-verify                                   → atl-verifier
→ /sdd-archive     (cierra issues + milestone)  → atl-verifier
```
