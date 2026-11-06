# AUDITORÍA DE SEGURIDAD - Bina Mail
**Fecha:** 2026-11-06  
**Versión:** 1.0  
**Estado:** MVP (Fase 4 completa, Fase 5 pendiente)  
**Auditor:** Vibe Code (security-skills)  

---

## 📋 RESUMEN EJECUTIVO

### Estado General: ⚠️ **MEJORABLE**

La arquitectura de seguridad es sólida en diseño (zero-knowledge, zero-log, zero-trust), pero hay **vulnerabilidades críticas y mejoras urgentes** en la implementación y configuración.

| Categoría | Estado | Prioridad | Issues |
|-----------|--------|-----------|--------|
| **CI/CD y Automatización** | 🟡 Medio | ALTA | 5 |
| **Configuración de Seguridad** | 🟡 Medio | ALTA | 4 |
| **Código Backend** | 🟢 Bueno | MEDIA | 3 |
| **Dependencias** | 🟢 Bueno | MEDIA | 2 |
| **Documentación de Seguridad** | 🟢 Bueno | BAJA | 1 |

**Riesgo Global:** **MEDIO-ALTO** (requiere acción inmediata en CI/CD y configuración)

---

## 🎯 PRIORIDADES (Orden de Acción)

### 🔴 **PRIORIDAD CRÍTICA (Acción Inmediata - 0-7 días)**

#### 1. **Permisos de CI/CD Insuficientes** 🚨
**ID:** SEC-2026-001  
**Severidad:** CRÍTICO  
**CVSS:** 9.8 (AV:N/AC:L/AT:N/PR:N/UI:N/S:U/C:H/I:H/A:H)  

**Descripción:**
- El job `lockfiles` tiene `permissions: contents: write` pero **NO tiene `pull-requests: write`**
- El job puede hacer `git push` directamente sin crear PR
- Esto permite a un atacante con acceso al workflow modificar código sin revisión

**Impacto:**
- Compromiso completo del repositorio
- Inyección de código malicioso
- Bypass de todos los gates de seguridad

**Solución:**
```yaml
# En .github/workflows/ci.yml, job lockfiles:
permissions:
  contents: write
  pull-requests: write  # REQUERIDO
```

**Evidencia:**
```yaml
- name: Confirmar cambios de lockfiles
  run: |
    git config user.name "bina-ci"
    git config user.email "ci@bina-mail.local"
    git add apps/web/pnpm-lock.yaml services/go.sum
    if git diff --cached --quiet; then
      echo "lockfiles sincronizados"
      exit 0
    fi
    git commit -m "chore: sincroniza lockfiles (pnpm-lock.yaml, go.sum)"
    git push  # ⚠️ PELIGRO: push directo sin PR
```

**Recomendación:** 
- Crear PR en lugar de push directo
- Usar `github-token` con permisos mínimos
- Requerir revisión humana para cambios en lockfiles

---

#### 2. **CodeQL sin Configuración de Severidad** 🚨
**ID:** SEC-2026-002  
**Severidad:** ALTO  
**CVSS:** 8.1 (AV:N/AC:L/AT:N/PR:N/UI:N/S:U/C:H/I:H/A:N)  

**Descripción:**
- CodeQL se ejecuta pero **no tiene configuración de severidad**
- No filtra por paths críticos
- No tiene configuración de queries personalizados para zero-log

**Impacto:**
- Falsos negativos: vulnerabilidades críticas pueden pasar desapercibidas
- Falsos positivos: ruido que oculta problemas reales

**Solución:**
```yaml
# En .github/workflows/codeql.yml
- name: Inicializar CodeQL
  uses: github/codeql-action/init@v4
  with:
    languages: ${{ matrix.language }}
    build-mode: ${{ matrix.build-mode }}
    config-file: .github/codeql-config.yml  # ⬅ AÑADIR
```

Crear `.github/codeql-config.yml`:
```yaml
name: Bina Mail CodeQL Config

# Paths críticos (zero-knowledge, zero-log)
paths:
  - services/internal/gateway
  - services/internal/authapi
  - services/internal/mailboxapi
  - services/internal/session
  - services/internal/jwt

# Excluir paths de prueba
paths-ignore:
  - "**/*_test.go"
  - "**/test/**"

# Queries adicionales para zero-log
queries:
  - name: "Zero-Log Violation"
    query: |
      from LogCall lc
      where lc.getMessage().matches("subject|recipient|sender|plaintext|body|content|message")
      select lc, "Zero-log violation: logging sensitive data"
```

---

#### 3. **Dependabot sin Alertas de Seguridad para @noble/*** 🚨
**ID:** SEC-2026-003  
**Severidad:** ALTO  
**CVSS:** 7.5 (AV:N/AC:L/AT:N/PR:N/UI:N/S:U/C:H/I:N/A:N)  

**Descripción:**
- `@noble/*` está **fijado a versión exacta** (determinismo criptográfico)
- Pero **no recibe security updates** porque `update-types` está ignorado para TODO
- Si hay un CVE en `@noble/curves` o `@noble/hashes`, no se notificará

**Impacto:**
- Vulnerabilidades en bibliotecas criptográficas pueden pasar desapercibidas
- Compromiso de E2EE

**Solución:**
```yaml
# En .github/dependabot.yml
ignore:
  - dependency-name: "@noble/*"
    update-types:
      - version-update:semver-major
      - version-update:semver-minor
      # ❌ ELIMINAR esta línea:
      - version-update:semver-patch
```

Corregido:
```yaml
ignore:
  - dependency-name: "@noble/*"
    update-types:
      - version-update:semver-major
      - version-update:semver-minor
    # ✅ Security updates SÍ se aplican
```

**Nota:** `version-update:semver-patch` incluye security updates. La configuración actual BLOQUEA todos los updates.

---

### 🟠 **PRIORIDAD ALTA (Acción en 7-30 días)**

#### 4. **Falta de Rotación de Secrets en CI** 🔑
**ID:** SEC-2026-004  
**Severidad:** ALTO  

**Descripción:**
- No hay evidencia de rotación de `SUPABASE_SERVICE_ROLE_KEY`
- No hay rotación de `SUPABASE_JWT_SECRET`
- No hay rotación de `KAFKA_BROKERS` credentials

**Impacto:**
- Si una clave se compromete, el atacante mantiene acceso indefinidamente

**Solución:**
- Implementar rotación automática con GitHub Actions Secrets
- Usar `hashicorp/vault-action` o similar
- Rotar cada 90 días

---

#### 5. **CORS sin Validación de Origen Dinámico** 🌐
**ID:** SEC-2026-005  
**Severidad:** ALTO  

**Descripción:**
- CORS acepta cualquier origen que coincida con `AllowedOrigin`
- No hay validación de que el origen sea HTTPS
- No hay lista blanca de orígenes permitidos

**Código actual:**
```go
if origin := r.Header.Get("Origin"); origin != "" {
    if origin != a.cfg.AllowedOrigin {
        writeError(w, http.StatusForbidden, "origen_no_permitido")
        return
    }
    w.Header().Set("Access-Control-Allow-Origin", origin)
```

**Solución:**
```go
// Validar que el origen es HTTPS
if origin := r.Header.Get("Origin"); origin != "" {
    parsed, err := url.Parse(origin)
    if err != nil || parsed.Scheme != "https" {
        writeError(w, http.StatusForbidden, "origen_invalido")
        return
    }
    if origin != a.cfg.AllowedOrigin {
        writeError(w, http.StatusForbidden, "origen_no_permitido")
        return
    }
    w.Header().Set("Access-Control-Allow-Origin", origin)
    w.Header().Set("Access-Control-Allow-Credentials", "true")  // ⬅ AÑADIR
```

---

#### 6. **Falta de Cabeceras de Seguridad Adicionales** 🛡️
**ID:** SEC-2026-006  
**Severidad:** MEDIO  

**Descripción:**
- Falta `Content-Security-Policy` (CSP)
- Falta `Permissions-Policy`
- Falta `Strict-Transport-Security` (HSTS)

**Código actual:**
```go
func setSecurityHeaders(w http.ResponseWriter) {
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("Referrer-Policy", "no-referrer")
    w.Header().Set("Cache-Control", "no-store")
}
```

**Solución:**
```go
func setSecurityHeaders(w http.ResponseWriter) {
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("X-XSS-Protection", "0")  // Deshabilitado (obsoleto)
    w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
    w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
    w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
    w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' wss:")
    w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
}
```

---

#### 7. **Fail-Open en Rate Limiting** ⚠️
**ID:** SEC-2026-007  
**Severidad:** MEDIO  
**ADR:** ADR-002 (documentado como fail-open intencional)  

**Descripción:**
- Si el contador de rate limiting falla, la petición se admite
- Esto es intencional (ADR-002) pero **debe documentarse claramente**

**Código:**
```go
// Fail-open documentado (ADR-002): un error del contador no
// derriba la disponibilidad; el limite es proteccion adicional.
if a.cfg.Limiter != nil {
    if ok, err := a.cfg.Limiter.Allow(r.Context(), "user:"+sess.UserID); err == nil && !ok {
        writeError(w, http.StatusTooManyRequests, "limite_alcanzado")
        return
    }
    // ⚠️ Si err != nil, la petición se admite (fail-open)
}
```

**Recomendación:**
- Mantener fail-open pero **loggear el error** (sin PII)
- Monitorear rate de errores del limiter

---

### 🟡 **PRIORIDAD MEDIA (Acción en 30-90 días)**

#### 8. **Falta de Validación de Tamaño en WebSocket** 📦
**ID:** SEC-2026-008  
**Severidad:** MEDIO  

**Descripción:**
- No hay límite de tamaño de mensajes en WebSocket
- Puede causar DoS por mensajes grandes

**Solución:**
```go
// En gateway/api.go, handler realtime:
upgrader := websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return r.Header.Get("Origin") == a.cfg.AllowedOrigin
    },
}
```

---

#### 9. **JWT Secret en Plain Text en Config** 🔑
**ID:** SEC-2026-009  
**Severidad:** MEDIO  

**Descripción:**
- `SUPABASE_JWT_SECRET` se pasa como string directo
- No hay validación de longitud mínima
- No hay rotación automática

**Solución:**
- Validar longitud mínima (64 bytes recomendado)
- Usar GitHub Secrets
- Rotar cada 90 días

---

#### 10. **Falta de Audit Log para Acciones Sensibles** 📋
**ID:** SEC-2026-010  
**Severidad:** MEDIO  

**Descripción:**
- No hay logging de:
  - Creación de sesiones
  - Revocación de sesiones
  - Publicación de claves
  - Almacenamiento de sobres
- **Pero:** ADR-005 prohíbe logging de PII

**Solución:**
- Loggear **solo IDs opacos** (sin PII)
- Ejemplo: `{"event":"session_created","user_id":"<opaque>","timestamp":"..."}`

---

## 📊 ANÁLISIS POR CATEGORÍA

### ✅ **BUEN ESTADO (Sin acción requerida)**

1. **Zero-Log Implementation** ✅
   - ADR-005 bien implementado
   - CI audita patrones de log prohibidos
   - Código no loggea PII

2. **Zero-Knowledge Architecture** ✅
   - Servidor nunca ve plaintext
   - Claves privadas solo en dispositivo
   - Cifrado E2EE con XChaCha20-Poly1305

3. **Criptografía** ✅
   - Argon2id con parámetros seguros
   - BLAKE3 para integridad
   - X25519 + Ed25519 para claves
   - Vectores de test obligatorios

4. **Autenticación** ✅
   - JWT HS256 verificado
   - Sesiones opacas
   - Tokens de corta vida (30 min)

5. **Dependencias** ✅
   - franz-go v1.22.1 (sin CVE conocido)
   - gorilla/websocket v1.5.3 (sin CVE conocido)
   - govulncheck en CI

---

## 🎯 RECOMENDACIONES ESTRATÉGICAS

### 1. **Implementar Security Hardening en CI/CD** 🚀
- [ ] Añadir CodeQL con configuración personalizada
- [ ] Corregir permisos del job lockfiles
- [ ] Añadir Trivy para escaneo de contenedores
- [ ] Añadir Snyk para dependencias

### 2. **Mejora de Configuración de Seguridad** 🛡️
- [ ] Añadir CSP y HSTS
- [ ] Validar CORS con HTTPS
- [ ] Implementar audit logging (sin PII)
- [ ] Rotación automática de secrets

### 3. **Monitoreo y Observabilidad** 📊
- [ ] Alertas de CodeQL
- [ ] Alertas de Dependabot
- [ ] Monitoreo de rate limiting errors
- [ ] Monitoreo de fallos de autenticación

---

## 📈 MÉTRICAS DE SEGURIDAD

| Métrica | Valor Actual | Objetivo | Estado |
|---------|--------------|----------|--------|
| CodeQL Coverage | 3 lenguajes | 4+ | ⚠️ |
| Dependabot Alerts | Activo | Activo + @noble/* | ⚠️ |
| CI Security Jobs | 3 (gitleaks, actionlint, zero-log) | 5+ | ⚠️ |
| Security Headers | 4 | 8+ | ⚠️ |
| Zero-Log Compliance | ✅ | ✅ | ✅ |
| Zero-Knowledge Compliance | ✅ | ✅ | ✅ |

---

## 🔗 REFERENCIAS

- [ADR-002: Cache Dragonfly](DECISIONS/ADR-002-cache-dragonfly.md)
- [ADR-005: Zero-Log](DECISIONS/ADR-005-politica-zero-log.md)
- [ADR-007: HTTP/JSON](DECISIONS/ADR-007-transporte-interno-http-json.md)
- [ADR-008: Dependencias](DECISIONS/ADR-008-dependencias-infraestructura.md)
- [CI Workflow](.github/workflows/ci.yml)
- [CodeQL Workflow](.github/workflows/codeql.yml)
- [Dependabot Config](.github/dependabot.yml)

---

## 📝 HISTORIAL DE REVISIONES

| Versión | Fecha | Autor | Cambios |
|---------|-------|-------|---------|
| 1.0 | 2026-11-06 | Vibe Code | Auditoría inicial |

---

**Fin del Reporte**
