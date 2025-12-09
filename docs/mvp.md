# Taskcore MVP - Feature Table

**Version:** 1.0
**Target:** 22 weeks (~5.5 months)
**Philosophy:** OpenJira - Self-hosted alternative to Atlassian Jira

---

## Complete MVP Feature Matrix

| Categoría                | Funcionalidad                      | Incluido en MVP | Prioridad | Detalles                                                         |
| ------------------------ | ---------------------------------- | :-------------: | :-------: | ---------------------------------------------------------------- |
| **Autenticación**        | Login clásico (usuario/contraseña) |        ✔        |    P0     | Contraseña con bcrypt; email único; login local opcional         |
| **Autenticación**        | Registro clásico                   |    ✔/opcional   |    P1     | Depende si MVP tiene invitaciones o registro libre               |
| **Autenticación**        | OpenID Connect (OIDC)              |   ✔ (opcional)  |    P2     | Keycloak/Auth0/Google; linking con cuenta local                  |
| **Autenticación**        | Linking de cuenta local ⇄ OIDC     |        ✔        |    P2     | Basado en email; configurable                                    |
| **Autenticación**        | Recuperación de contraseña         |        ✔        |    P1     | Token por email                                                  |
| **Autorización / Roles** | Admin                              |        ✔        |    P0     | Flag `is_admin` en users, acceso total al sistema                |
| **Autorización / Roles** | Project Admin                      |        ✔        |    P0     | Configura workflows, tipos de issue, miembros                    |
| **Autorización / Roles** | Developer                          |        ✔        |    P0     | CRUD issues, sub-tareas, comentarios                             |
| **Autorización / Roles** | Reporter                           |        ✔        |    P0     | Crea issues, comenta, no puede editar flujos                     |
| **Grupos**               | Grupos globales                    |        ✔        |    P0     | Colecciones globales de usuarios (taskcore-developers, etc.)     |
| **Grupos**               | Membresía de usuarios              |        ✔        |    P0     | Usuarios pertenecen a múltiples grupos                           |
| **Grupos**               | Asignación bulk a proyectos        |        ✔        |    P0     | Asignar grupo entero a un proyecto con un rol                    |
| **Permission Schemes**   | Permission schemes reusables       |        ✔        |    P0     | Sets de permisos reutilizables para proyectos                    |
| **Permission Schemes**   | Grantee types (4 tipos)            |        ✔        |    P0     | User, Group, Project Role, Anyone                                |
| **Permission Schemes**   | Default permission scheme          |        ✔        |    P0     | Scheme por defecto con permisos sensatos                         |
| **Permission Schemes**   | Custom permission schemes          |        ✔        |    P0     | Crear schemes personalizados para proyectos específicos          |
| **Tipos de Issue**       | Epic                               |        ✔        |    P0     | Nivel más alto                                                   |
| **Tipos de Issue**       | User Story                         |        ✔        |    P0     | Pertenecen a Epics; aceptan subtareas                            |
| **Tipos de Issue**       | Task                               |        ✔        |    P0     | Trabajo individual; admite subtareas                             |
| **Tipos de Issue**       | Sub-task                           |        ✔        |    P0     | Unidad mínima                                                    |
| **Tipos de Issue**       | Bug                                |        ✔        |    P0     | Workflow independiente                                           |
| **Workflows**            | Estados básicos                    |        ✔        |    P0     | `To Do → In Progress → In Review → Done`                         |
| **Workflows**            | Flujos distintos por tipo de issue |        ✔        |    P0     | Ej.: `Bug: New → Triaged → Fixing → QA → Done`                   |
| **Workflows**            | Reglas simples                     |        ✔        |    P0     | Quién puede mover cada estado                                    |
| **Workflows**            | Transiciones personalizables       |        ✔        |    P1     | UI simple; no full engine como Jira                              |
| **Proyectos**            | Crear proyectos                    |        ✔        |    P0     | Opciones: Kanban, Scrum                                          |
| **Proyectos**            | Configurar tipos de issue activos  |        ✔        |    P1     | Ej.: desactivar "Epic" en un proyecto simple                     |
| **Proyectos**            | Miembros + roles por proyecto      |        ✔        |    P0     | Admin, Developer, Reporter                                       |
| **Board Kanban**         | Columnas (estados)                 |        ✔        |    P0     | Basado en workflow activo                                        |
| **Board Kanban**         | Arrastrar y soltar                 |        ✔        |    P0     | Drag & drop                                                      |
| **Board Kanban**         | Filtros básicos                    |        ✔        |    P0     | Por assignee, tipo, estado                                       |
| **Scrum**                | Backlog                            |        ✔        |    P0     | Lista ordenada                                                   |
| **Scrum**                | Sprint básico                      |        ✔        |    P0     | Create/start/close sprint                                        |
| **Scrum**                | Velocity simple                    |        ✔        |    P1     | Story points opcional                                            |
| **Scrum**                | Burndown                           |        ✔        |    P1     | Gráfico básico                                                   |
| **Issues**               | CRUD completo                      |        ✔        |    P0     | Crear, editar, ver historial                                     |
| **Issues**               | Asignación                         |        ✔        |    P0     | 1 responsable                                                    |
| **Issues**               | Comentarios                        |        ✔        |    P0     | Soporta markdown                                                 |
| **Issues**               | Adjuntos                           |    ✔/opcional   |    P2     | S3/minio o disco local                                           |
| **Issues**               | Tags / labels                      |        ✔        |    P0     | Colores simples                                                  |
| **Issues**               | Story points                       |    ✔/opcional   |    P1     | Para Scrum                                                       |
| **Issues**               | Relación entre issues              |        ✔        |    P1     | Blocked by / Blocks / Relates                                    |
| **JQL-like (TQL)**       | Búsqueda simple                    |        ✔        |    P0     | `status = "In Progress"`                                         |
| **JQL-like (TQL)**       | Filtros combinados                 |        ✔        |    P0     | `assignee = julio AND type = bug`                                |
| **JQL-like (TQL)**       | OR / AND                           |        ✔        |    P0     | Soporte básico                                                   |
| **JQL-like (TQL)**       | Ordenamientos                      |        ✔        |    P0     | `ORDER BY created_at DESC`                                       |
| **Notificaciones**       | Email básico                       |        ✔        |    P1     | Cambios de estado / asignación                                   |
| **Integraciones**        | Webhooks                           |    ✔/opcional   |    P2     | Solo project-level                                               |
| **Personalización**      | Tema claro/oscuro                  |        ✔        |    P1     | switch                                                           |
| **Personalización**      | Logo del proyecto                  |        ✔        |    P2     | opcional                                                         |
| **Infraestructura**      | Un solo binario back+front         |        ✔        |    P0     | Similar a Gitea (Go + React embebido)                            |
| **Infraestructura**      | Base de datos                      |        ✔        |    P0     | PostgreSQL recomendado                                           |
| **Infraestructura**      | Migraciones automáticas            |        ✔        |    P0     | Con `golang-migrate`                                             |
| **Infraestructura**      | Configuración YAML o env           |        ✔        |    P0     | `config.yaml` o variables                                        |

---

## Prioridades

**P0 (Must Have):** Core del MVP, sin esto no funciona
**P1 (Should Have):** Importante pero puede posponerse 1-2 semanas
**P2 (Nice to Have):** Mejoras que pueden ir en post-MVP

---

## Simplificaciones vs Jira

Para mantener el MVP simple, **no incluimos** (pero pueden agregarse post-MVP):

- **Permission Schemes complejos:** Solo roles simples (Admin, Developer, Reporter)
- **Workflows con condiciones complejas:** Solo transiciones básicas con roles
- **Custom fields:** Campos fijos en el MVP
- **Issue security levels:** Todo visible dentro del proyecto
- **Notification schemes:** Emails simples predefinidos
- **Time tracking avanzado:** Solo worklogs básicos (opcional P2)
- **Roadmaps / Gantt:** Solo backlog y sprints
- **Multi-tenant SaaS:** Solo self-hosted

---

## Next Steps

1. ✅ Diseñar schema SQL completo → `docs/schema.md`
2. ✅ Documentar sistema de permisos → `docs/permissions.md`
3. [ ] Crear migraciones SQL incrementales
4. [ ] Implementar modelos Go en `internal/repository/models/`
5. [ ] Crear API REST básica en `internal/handler/`
